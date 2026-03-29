package cargo

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// CreateCargoUseCase registers a new cargo shipment assigned to a vessel.
type CreateCargoUseCase struct {
	repo CargoRepository
}

// NewCreateCargoUseCase creates a new use case with injected dependencies.
func NewCreateCargoUseCase(repo CargoRepository) *CreateCargoUseCase {
	return &CreateCargoUseCase{repo: repo}
}

// Execute creates a new cargo.
func (uc *CreateCargoUseCase) Execute(ctx context.Context, input domaincargo.CreateCargoInput) (*domaincargo.Cargo, error) {
	// Input validation is performed by domain model
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Persist cargo to repository (orchestration only, no business logic)
	cargo, err := uc.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create_cargo: %w", err)
	}

	// Log successful creation
	zerolog.Ctx(ctx).Info().
		Str("cargo_id", cargo.ID.String()).
		Str("name", cargo.Name).
		Msg("cargo created")

	return cargo, nil
}
