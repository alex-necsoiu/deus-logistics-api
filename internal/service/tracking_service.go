package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// TrackingService implements tracking.Service.
type TrackingService struct {
	repo tracking.Repository
}

// NewTrackingService creates a new tracking service with the given repository.
//
// Inputs:
//   repo - tracking repository implementation (must not be nil)
//
// Returns:
//   *TrackingService with initialized repository
func NewTrackingService(repo tracking.Repository) *TrackingService {
	return &TrackingService{repo: repo}
}

// AddTrackingEntry creates a new tracking entry for a cargo shipment.
//
// Inputs:
//   ctx   - request context for cancellation and tracing
//   input - tracking entry details (CargoID, Location, Status required)
//
// Returns:
//   *TrackingEntry on success
//   ErrInvalidEntry if CargoID is nil, Location is empty, or Status is empty
//
// Side effects:
//   - DB write to tracking_entries table
//   - Logs info message with cargo_id and status
func (s *TrackingService) AddTrackingEntry(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
	entry, err := s.repo.Append(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("addTrackingEntry: %w", err)
	}
	zerolog.Ctx(ctx).Info().Str("cargo_id", input.CargoID.String()).Str("status", input.Status).Msg("tracking entry appended")
	return entry, nil
}

// GetTrackingHistory retrieves all tracking entries for a specific cargo shipment.
//
// Inputs:
//   ctx    - request context for cancellation and tracing
//   cargoID - UUID of the cargo (must not be nil)
//
// Returns:
//   []*TrackingEntry sorted by timestamp on success
//   Empty slice if no tracking entries exist
//   ErrInvalidEntry if cargoID is nil
//
// Side effects:
//   - DB read from tracking_entries table
func (s *TrackingService) GetTrackingHistory(ctx context.Context, cargoID uuid.UUID) ([]*tracking.TrackingEntry, error) {
	if cargoID == uuid.Nil {
		return nil, tracking.ErrInvalidEntry
	}
	entries, err := s.repo.ListByCargoID(ctx, cargoID)
	if err != nil {
		return nil, fmt.Errorf("getTrackingHistory: %w", err)
	}
	if entries == nil {
		entries = []*tracking.TrackingEntry{}
	}
	return entries, nil
}
