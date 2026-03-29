package cargo

import (
	"context"
	"fmt"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// ListCargosUseCase retrieves all cargo records.
type ListCargosUseCase struct {
	repo CargoRepository
}

// NewListCargosUseCase creates a new use case with injected dependencies.
func NewListCargosUseCase(repo CargoRepository) *ListCargosUseCase {
	return &ListCargosUseCase{repo: repo}
}

// Execute retrieves all cargos.
func (uc *ListCargosUseCase) Execute(ctx context.Context) ([]*domaincargo.Cargo, error) {
	// Fetch cargos from repository (orchestration only)
	cargos, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list_cargos: %w", err)
	}

	// Ensure non-nil slice for API responses
	if cargos == nil {
		cargos = []*domaincargo.Cargo{}
	}

	return cargos, nil
}
