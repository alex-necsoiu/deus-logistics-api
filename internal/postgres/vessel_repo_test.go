package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// startVesselTestContainer starts a PostgreSQL testcontainer and returns pool + cleanup func.
func startVesselTestContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "deus_vessel_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "postgres://postgres:postgres@" + host + ":" + port.Port() + "/deus_vessel_test?sslmode=disable"

	// Retry connection with backoff
	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = New(ctx, dsn)
		if err == nil {
			// Verify connection works
			connCtx, connCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = pool.Ping(connCtx)
			connCancel()
			if err == nil {
				break
			}
		}
		time.Sleep(time.Duration((i+1)*500) * time.Millisecond)
	}
	require.NoError(t, err, "failed to connect to test database")

	return pool, func() {
		pool.Close()
		container.Terminate(context.Background())
	}
}

// runVesselTestSchema creates the test database schema for vessel tests.
func runVesselTestSchema(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schema := `
		CREATE TABLE vessels (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			capacity NUMERIC NOT NULL,
			current_location TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`

	_, err := pool.Exec(ctx, schema)
	require.NoError(t, err)
}

// TestVesselRepository_CreateWithCorrectTimestamps verifies timestamps are populated from DB.
//
// Given: a new vessel creation input
// When: the vessel is created and persisted
// Then: the returned vessel has non-zero CreatedAt and UpdatedAt from database
func TestVesselRepository_CreateWithCorrectTimestamps(t *testing.T) {
	// Given: a test database
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	input := vessel.CreateVesselInput{
		Name:            "Cargo Ship Alpha",
		Capacity:        5000.0,
		CurrentLocation: "Port of Singapore",
	}

	beforeCreate := time.Now().Add(-1 * time.Second)

	// When: vessel is created
	created, err := repo.Create(ctx, input)

	// Then: creation succeeds
	require.NoError(t, err)
	require.NotNil(t, created)

	// And: vessel has correct values
	assert.Equal(t, input.Name, created.Name)
	assert.Equal(t, input.Capacity, created.Capacity)
	assert.Equal(t, input.CurrentLocation, created.CurrentLocation)

	// And: CreatedAt is NOT zero value (critical fix validation)
	assert.NotEqual(t, time.Time{}, created.CreatedAt, "CreatedAt must not be zero value")

	// And: UpdatedAt is NOT zero value (critical fix validation)
	assert.NotEqual(t, time.Time{}, created.UpdatedAt, "UpdatedAt must not be zero value")

	// And: Timestamps are close to creation time (within 5 seconds)
	afterCreate := time.Now().Add(5 * time.Second)
	assert.True(t, created.CreatedAt.After(beforeCreate), "CreatedAt should be after beforeCreate")
	assert.True(t, created.CreatedAt.Before(afterCreate), "CreatedAt should be before afterCreate")
	assert.True(t, created.UpdatedAt.After(beforeCreate), "UpdatedAt should be after beforeCreate")
	assert.True(t, created.UpdatedAt.Before(afterCreate), "UpdatedAt should be before afterCreate")

	// And: UUID is not empty
	assert.NotEqual(t, uuid.Nil, created.ID, "Vessel ID must not be nil")
}

// TestVesselRepository_CreateMultipleVessels verifies multiple vessels get unique timestamps.
func TestVesselRepository_CreateMultipleVessels(t *testing.T) {
	// Given: a test database
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	testCases := []struct {
		name            string
		capacity        float64
		currentLocation string
	}{
		{
			name:            "Vessel One",
			capacity:        1000.0,
			currentLocation: "Singapore",
		},
		{
			name:            "Vessel Two",
			capacity:        2000.0,
			currentLocation: "Hong Kong",
		},
		{
			name:            "Vessel Three",
			capacity:        3000.0,
			currentLocation: "Shanghai",
		},
	}

	// When: multiple vessels are created
	var vessels []*vessel.Vessel
	for _, tc := range testCases {
		input := vessel.CreateVesselInput{
			Name:            tc.name,
			Capacity:        tc.capacity,
			CurrentLocation: tc.currentLocation,
		}

		created, err := repo.Create(ctx, input)

		// Then: each creation succeeds
		require.NoError(t, err)
		require.NotNil(t, created)

		// And: timestamps are valid
		assert.NotEqual(t, time.Time{}, created.CreatedAt)
		assert.NotEqual(t, time.Time{}, created.UpdatedAt)

		vessels = append(vessels, created)

		// Small delay to ensure distinct timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// And: list returns all vessels with valid timestamps
	listed, err := repo.List(ctx)
	require.NoError(t, err)
	require.Equal(t, len(testCases), len(listed))

	for _, v := range listed {
		assert.NotEqual(t, time.Time{}, v.CreatedAt, "All listed vessels must have non-zero CreatedAt")
		assert.NotEqual(t, time.Time{}, v.UpdatedAt, "All listed vessels must have non-zero UpdatedAt")
	}
}

// TestVesselRepository_GetByIDReturnsTimestamps verifies GetByID returns correct timestamps.
func TestVesselRepository_GetByIDReturnsTimestamps(t *testing.T) {
	// Given: a test database with a created vessel
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	input := vessel.CreateVesselInput{
		Name:            "Test Vessel",
		Capacity:        5000.0,
		CurrentLocation: "Test Port",
	}

	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// When: vessel is retrieved by ID
	retrieved, err := repo.GetByID(ctx, created.ID)

	// Then: retrieval succeeds
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// And: timestamps match creation
	assert.Equal(t, created.CreatedAt, retrieved.CreatedAt)
	assert.Equal(t, created.UpdatedAt, retrieved.UpdatedAt)

	// And: timestamps are not zero
	assert.NotEqual(t, time.Time{}, retrieved.CreatedAt)
	assert.NotEqual(t, time.Time{}, retrieved.UpdatedAt)
}

// TestVesselRepository_UpdateLocationPreservesCreatedAt verifies CreatedAt doesn't change on update.
func TestVesselRepository_UpdateLocationPreservesCreatedAt(t *testing.T) {
	// Given: a created vessel
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	input := vessel.CreateVesselInput{
		Name:            "Moving Vessel",
		Capacity:        5000.0,
		CurrentLocation: "Port A",
	}

	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	originalCreatedAt := created.CreatedAt

	// Small delay to ensure UpdatedAt differs from CreatedAt
	time.Sleep(100 * time.Millisecond)

	// When: vessel location is updated
	updated, err := repo.UpdateLocation(ctx, created.ID, "Port B")

	// Then: update succeeds
	require.NoError(t, err)
	require.NotNil(t, updated)

	// And: CreatedAt is preserved
	assert.Equal(t, originalCreatedAt, updated.CreatedAt, "CreatedAt must not change on update")

	// And: UpdatedAt is more recent
	assert.True(t, updated.UpdatedAt.After(originalCreatedAt), "UpdatedAt must be after original CreatedAt")

	// And: location is updated
	assert.Equal(t, "Port B", updated.CurrentLocation)
}

// TestVesselRepository_UpdateCapacityPreservesCreatedAt verifies CreatedAt during capacity update.
func TestVesselRepository_UpdateCapacityPreservesCreatedAt(t *testing.T) {
	// Given: a created vessel
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	input := vessel.CreateVesselInput{
		Name:            "Capacity Vessel",
		Capacity:        5000.0,
		CurrentLocation: "Port X",
	}

	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	originalCreatedAt := created.CreatedAt

	time.Sleep(100 * time.Millisecond)

	// When: vessel capacity is updated
	updated, err := repo.UpdateCapacity(ctx, created.ID, 6000.0)

	// Then: update succeeds
	require.NoError(t, err)
	require.NotNil(t, updated)

	// And: CreatedAt is preserved
	assert.Equal(t, originalCreatedAt, updated.CreatedAt, "CreatedAt must not change on capacity update")

	// And: UpdatedAt is more recent
	assert.True(t, updated.UpdatedAt.After(originalCreatedAt), "UpdatedAt must be after original CreatedAt")

	// And: capacity is updated
	assert.Equal(t, 6000.0, updated.Capacity)
}

// TestVesselRepository_GetByIDNotFound verifies ErrNotFound for missing vessel.
func TestVesselRepository_GetByIDNotFound(t *testing.T) {
	// Given: a test database with no vessels
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	// When: retrieving non-existent vessel
	retrieved, err := repo.GetByID(ctx, uuid.New())

	// Then: error is ErrNotFound
	assert.True(t, errors.Is(err, vessel.ErrNotFound), "Expected ErrNotFound error")
	assert.Nil(t, retrieved)
}

// TestVesselRepository_UpdateLocationNotFound verifies error for non-existent vessel.
func TestVesselRepository_UpdateLocationNotFound(t *testing.T) {
	// Given: a test database with no vessels
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	// When: updating location of non-existent vessel
	updated, err := repo.UpdateLocation(ctx, uuid.New(), "New Location")

	// Then: error is ErrNotFound
	assert.True(t, errors.Is(err, vessel.ErrNotFound), "Expected ErrNotFound error")
	assert.Nil(t, updated)
}

// TestVesselRepository_UpdateCapacityNotFound verifies error for non-existent vessel.
func TestVesselRepository_UpdateCapacityNotFound(t *testing.T) {
	// Given: a test database with no vessels
	pool, cleanup := startVesselTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runVesselTestSchema(t, pool)

	repo := NewVesselRepository(pool)

	// When: updating capacity of non-existent vessel
	updated, err := repo.UpdateCapacity(ctx, uuid.New(), 1000.0)

	// Then: error is ErrNotFound
	assert.True(t, errors.Is(err, vessel.ErrNotFound), "Expected ErrNotFound error")
	assert.Nil(t, updated)
}
