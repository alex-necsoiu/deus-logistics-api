package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// startTestContainer starts a PostgreSQL testcontainer and returns pool + cleanup func.
func startTestContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "deus_test",
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

	dsn := "postgres://postgres:postgres@" + host + ":" + port.Port() + "/deus_test?sslmode=disable"

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
		time.Sleep(time.Duration((i + 1) * 500) * time.Millisecond)
	}
	require.NoError(t, err, "failed to connect to test database")

	return pool, func() {
		pool.Close()
		container.Terminate(context.Background())
	}
}

// runTestSchema creates the test database schema.
func runTestSchema(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schema := `
		CREATE TABLE vessels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			capacity NUMERIC NOT NULL,
			current_location TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE cargoes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			description TEXT,
			weight NUMERIC NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('pending', 'in_transit', 'delivered')),
			vessel_id UUID NOT NULL REFERENCES vessels(id) ON DELETE RESTRICT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX idx_cargoes_vessel_id ON cargoes(vessel_id);
		CREATE INDEX idx_cargoes_status ON cargoes(status);
	`

	_, err := pool.Exec(ctx, schema)
	require.NoError(t, err)
}

// TestCargoRepository_CreatePersistsDescription verifies cargo descriptions are persisted.
func TestCargoRepository_CreatePersistsDescription(t *testing.T) {
	// Given: a test database
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create test vessel
	var vesselID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Test Vessel", 1000.0, "Port of Origin",
	).Scan(&vesselID)
	require.NoError(t, err)

	input := cargo.CreateCargoInput{
		Name:        "Electronics Shipment",
		Description: "Fragile goods - handle with care",
		Weight:      150.5,
		VesselID:    vesselID,
	}

	// When: cargo is created
	created, err := repo.Create(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Then: description is returned immediately
	assert.Equal(t, input.Description, created.Description)
	assert.Equal(t, input.Name, created.Name)
	assert.Equal(t, input.Weight, created.Weight)
	assert.Equal(t, cargo.CargoStatusPending, created.Status)

	// And: when retrieved, description persists
	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, input.Description, retrieved.Description)
}

// TestCargoRepository_GetByIDReturnsDescription verifies GetByID retrieves description.
func TestCargoRepository_GetByIDReturnsDescription(t *testing.T) {
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create test vessel
	var vesselID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Test Vessel", 1000.0, "Port of Origin",
	).Scan(&vesselID)
	require.NoError(t, err)

	// Given: a cargo with description
	created, err := repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Test Item",
		Description: "This is a detailed description",
		Weight:      50.0,
		VesselID:    vesselID,
	})
	require.NoError(t, err)

	// When: cargo is retrieved by ID
	retrieved, err := repo.GetByID(ctx, created.ID)

	// Then: description is present
	require.NoError(t, err)
	assert.Equal(t, "This is a detailed description", retrieved.Description)
}

// TestCargoRepository_ListIncludesDescription verifies List returns descriptions.
func TestCargoRepository_ListIncludesDescription(t *testing.T) {
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create test vessel
	var vesselID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Test Vessel", 1000.0, "Port of Origin",
	).Scan(&vesselID)
	require.NoError(t, err)

	// Given: multiple cargoes with descriptions
	cargo1, err := repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Cargo 1",
		Description: "Description 1",
		Weight:      100.0,
		VesselID:    vesselID,
	})
	require.NoError(t, err)

	cargo2, err := repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Cargo 2",
		Description: "Description 2",
		Weight:      200.0,
		VesselID:    vesselID,
	})
	require.NoError(t, err)

	// When: all cargoes are listed
	cargoes, err := repo.List(ctx)

	// Then: descriptions are included
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(cargoes), 2)

	// Verify both descriptions are present
	found1 := false
	found2 := false
	for _, c := range cargoes {
		if c.ID == cargo1.ID {
			assert.Equal(t, "Description 1", c.Description)
			found1 = true
		}
		if c.ID == cargo2.ID {
			assert.Equal(t, "Description 2", c.Description)
			found2 = true
		}
	}
	assert.True(t, found1, "Cargo 1 not found in list")
	assert.True(t, found2, "Cargo 2 not found in list")
}

// TestCargoRepository_ListByVesselIDIncludesDescription verifies descriptions in vessel cargo list.
func TestCargoRepository_ListByVesselIDIncludesDescription(t *testing.T) {
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create two test vessels
	var vessel1, vessel2 uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Vessel 1", 1000.0, "Port 1",
	).Scan(&vessel1)
	require.NoError(t, err)

	err = pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Vessel 2", 2000.0, "Port 2",
	).Scan(&vessel2)
	require.NoError(t, err)

	// Given: cargoes for different vessels
	cargo1, err := repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Vessel 1 Cargo",
		Description: "Only this description should appear",
		Weight:      100.0,
		VesselID:    vessel1,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Vessel 2 Cargo",
		Description: "This should not appear",
		Weight:      200.0,
		VesselID:    vessel2,
	})
	require.NoError(t, err)

	// When: cargoes for vessel 1 are listed
	cargoes, err := repo.ListByVesselID(ctx, vessel1)

	// Then: only vessel 1's cargo with correct description is returned
	require.NoError(t, err)
	require.Equal(t, 1, len(cargoes))
	assert.Equal(t, cargo1.ID, cargoes[0].ID)
	assert.Equal(t, "Only this description should appear", cargoes[0].Description)
}

// TestCargoRepository_UpdateStatusPreservesDescription verifies description survives status updates.
func TestCargoRepository_UpdateStatusPreservesDescription(t *testing.T) {
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create test vessel
	var vesselID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Test Vessel", 1000.0, "Port of Origin",
	).Scan(&vesselID)
	require.NoError(t, err)

	// Given: a cargo with description
	created, err := repo.Create(ctx, cargo.CreateCargoInput{
		Name:        "Status Test Cargo",
		Description: "This description must persist through status changes",
		Weight:      75.0,
		VesselID:    vesselID,
	})
	require.NoError(t, err)

	// When: status is updated
	updated, err := repo.UpdateStatus(ctx, created.ID, cargo.CargoStatusInTransit)

	// Then: description is preserved
	require.NoError(t, err)
	assert.Equal(t, "This description must persist through status changes", updated.Description)
	assert.Equal(t, cargo.CargoStatusInTransit, updated.Status)
}

// TestCargoRepository_EmptyDescriptionHandled verifies empty descriptions work.
func TestCargoRepository_EmptyDescriptionHandled(t *testing.T) {
	pool, cleanup := startTestContainer(t)
	defer cleanup()

	ctx := context.Background()
	runTestSchema(t, pool)

	repo := NewCargoRepository(pool)

	// Create test vessel
	var vesselID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO vessels (name, capacity, current_location) VALUES ($1, $2, $3) RETURNING id`,
		"Test Vessel", 1000.0, "Port of Origin",
	).Scan(&vesselID)
	require.NoError(t, err)

	// Given: cargo with empty description
	input := cargo.CreateCargoInput{
		Name:        "No Description Cargo",
		Description: "",
		Weight:      25.0,
		VesselID:    vesselID,
	}

	// When: cargo is created
	created, err := repo.Create(ctx, input)
	require.NoError(t, err)

	// Then: empty description is handled correctly
	assert.Equal(t, "", created.Description)

	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "", retrieved.Description)
}
