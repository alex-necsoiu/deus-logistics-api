package cargo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// GetCargoUseCase retrieves a cargo by ID.
type GetCargoUseCase struct {
	repo CargoRepository
}

// NewGetCargoUseCase creates a new use case with injected dependencies.
func NewGetCargoUseCase(repo CargoRepository) *GetCargoUseCase {
	return &GetCargoUseCase{repo: repo}
}

// Execute retrieves a cargo by ID.
func (uc *GetCargoUseCase) Execute(ctx context.Context, id uuid.UUID) (*domaincargo.Cargo, error) {
	// Validation: ID must not be nil
	if id == uuid.Nil {
		return nil, domaincargo.ErrInvalidInput
	}

	// Fetch cargo from repository (orchestration only)
	cargo, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get_cargo: %w", err)
	}

	return cargo, nil
}
