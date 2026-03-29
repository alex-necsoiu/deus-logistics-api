package health

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents the result of a health check
type Check struct {
	Name   string        `json:"name"`
	Status Status        `json:"status"`
	Error  string        `json:"error,omitempty"`
	Time   time.Duration `json:"response_time_ms"`
}

// Reporter checks health of various components
type Reporter struct {
	db *pgxpool.Pool
}

// NewReporter creates a new health reporter
func NewReporter(db *pgxpool.Pool) *Reporter {
	return &Reporter{db: db}
}

// CheckLiveness verifies the API is running (basic health)
// Returns 200 if the API process is alive
func (r *Reporter) CheckLiveness(ctx context.Context) Check {
	start := time.Now()
	return Check{
		Name:   "liveness",
		Status: StatusHealthy,
		Time:   time.Since(start),
	}
}

// CheckDatabase verifies database connectivity and responsiveness
func (r *Reporter) CheckDatabase(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name: "database",
	}

	if r.db == nil {
		check.Status = StatusUnhealthy
		check.Error = "database connection not initialized"
		check.Time = time.Since(start)
		return check
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.Ping(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Error = err.Error()
	} else {
		check.Status = StatusHealthy
	}

	check.Time = time.Since(start)
	return check
}

// CheckReadiness performs comprehensive readiness checks
// Returns checks for all critical dependencies
type ReadinessResult struct {
	Ready   bool    `json:"ready"`
	Status  Status  `json:"status"`
	Checks  []Check `json:"checks"`
	Version string  `json:"version,omitempty"`
}

// CheckReadiness checks if the application is ready to serve traffic
func (r *Reporter) CheckReadiness(ctx context.Context) ReadinessResult {
	checks := []Check{
		r.CheckDatabase(ctx),
	}

	// Determine overall readiness
	allHealthy := true
	for _, check := range checks {
		if check.Status != StatusHealthy {
			allHealthy = false
			break
		}
	}

	return ReadinessResult{
		Ready:  allHealthy,
		Status: StatusHealthy,
		Checks: checks,
	}
}
