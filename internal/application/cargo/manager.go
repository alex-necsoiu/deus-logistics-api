package cargo

// CargoApplicationManager manages all cargo use cases.
// It provides a single point of dependency injection for the application layer.
type CargoApplicationManager struct {
	CreateCargo        *CreateCargoUseCase
	GetCargo           *GetCargoUseCase
	ListCargos         *ListCargosUseCase
	ListCargosByVessel *ListCargosByVesselUseCase
	UpdateStatus       *UpdateCargoStatusUseCase
}

// NewCargoApplicationManager creates and wires all cargo use cases.
func NewCargoApplicationManager(
	cargoRepo CargoRepository,
	trackingRepo TrackingRepository,
	publisher EventPublisher,
) *CargoApplicationManager {
	return &CargoApplicationManager{
		CreateCargo:        NewCreateCargoUseCase(cargoRepo),
		GetCargo:           NewGetCargoUseCase(cargoRepo),
		ListCargos:         NewListCargosUseCase(cargoRepo),
		ListCargosByVessel: NewListCargosByVesselUseCase(cargoRepo),
		UpdateStatus:       NewUpdateCargoStatusUseCase(cargoRepo, trackingRepo, publisher),
	}
}
