package cargo

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// ListCargosByVesselUseCase retrieves all cargo records for a specific vessel.
type ListCargosByVesselUseCase struct {
	repo CargoRepository
}

// NewListCargosByVesselUseCase creates a new use case with injected dependencies.
func NewListCargosByVesselUseCase(repo CargoRepository) *ListCargosByVesselUseCase {
	return &ListCargosByVesselUseCase{repo: repo}
}

// Execute retrieves cargos for a specific vessel.
func (uc *ListCargosByVesselUseCase) Execute(ctx context.Context, vesselID uuid.UUID) ([]*domaincargo.Cargo, error) {
	// Validation: vessel ID must not be nil
	if vesselID == uuid.Nil {
		return nil, domaincargo.ErrInvalidInput
	}

	// Fetch cargos from repository (orchestration only)
	cargos, err := uc.repo.ListByVesselID(ctx, vesselID)
	if err != nil {
		return nil, fmt.Errorf("list_cargos_by_vessel: %w", err)
	}

	// Ensure non-nil slice for API responses
	if cargos == nil {
		cargos = []*domaincargo.Cargo{}
	}

	return cargos, nil
}
