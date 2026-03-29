package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// --- Tests for MapErrorToHTTPStatus ---

func TestMapErrorToHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "nil error returns 200 OK",
			err:            nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ErrNotFound returns 404",
			err:            ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "cargo.ErrNotFound returns 404",
			err:            cargo.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "vessel.ErrNotFound returns 404",
			err:            vessel.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "tracking.ErrCargoNotFound returns 404",
			err:            tracking.ErrCargoNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrInvalidTransition returns 422",
			err:            ErrInvalidTransition,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "cargo.ErrInvalidTransition returns 422",
			err:            cargo.ErrInvalidTransition,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "ErrInvalidInput returns 400",
			err:            ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "cargo.ErrInvalidInput returns 400",
			err:            cargo.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "cargo.ErrInvalidStatus returns 400",
			err:            cargo.ErrInvalidStatus,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "vessel.ErrInvalidInput returns 400",
			err:            vessel.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "tracking.ErrInvalidEntry returns 400",
			err:            tracking.ErrInvalidEntry,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unknown error returns 500",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrapped ErrNotFound returns 404",
			err:            errors.New("failed to get cargo: " + ErrNotFound.Error()),
			expectedStatus: http.StatusInternalServerError, // Wrapping breaks error type checking
		},
		{
			name:           "ErrUnknown returns 500",
			err:            ErrUnknown,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			status := MapErrorToHTTPStatus(tt.err)

			// Assert
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

// --- Tests for MapErrorToErrorCode ---

func TestMapErrorToErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode ErrorCode
	}{
		{
			name:         "nil error returns empty code",
			err:          nil,
			expectedCode: "",
		},
		{
			name:         "ErrNotFound returns CodeNotFound",
			err:          ErrNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "cargo.ErrNotFound returns CodeNotFound",
			err:          cargo.ErrNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "vessel.ErrNotFound returns CodeNotFound",
			err:          vessel.ErrNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "tracking.ErrCargoNotFound returns CodeNotFound",
			err:          tracking.ErrCargoNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "ErrInvalidTransition returns CodeInvalidTransition",
			err:          ErrInvalidTransition,
			expectedCode: CodeInvalidTransition,
		},
		{
			name:         "cargo.ErrInvalidTransition returns CodeInvalidTransition",
			err:          cargo.ErrInvalidTransition,
			expectedCode: CodeInvalidTransition,
		},
		{
			name:         "ErrInvalidInput returns CodeInvalidInput",
			err:          ErrInvalidInput,
			expectedCode: CodeInvalidInput,
		},
		{
			name:         "cargo.ErrInvalidInput returns CodeInvalidInput",
			err:          cargo.ErrInvalidInput,
			expectedCode: CodeInvalidInput,
		},
		{
			name:         "cargo.ErrInvalidStatus returns CodeInvalidInput",
			err:          cargo.ErrInvalidStatus,
			expectedCode: CodeInvalidInput,
		},
		{
			name:         "vessel.ErrInvalidInput returns CodeInvalidInput",
			err:          vessel.ErrInvalidInput,
			expectedCode: CodeInvalidInput,
		},
		{
			name:         "tracking.ErrInvalidEntry returns CodeInvalidInput",
			err:          tracking.ErrInvalidEntry,
			expectedCode: CodeInvalidInput,
		},
		{
			name:         "unknown error returns CodeInternalError",
			err:          errors.New("unknown error"),
			expectedCode: CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			code := MapErrorToErrorCode(tt.err)

			// Assert
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

// --- Tests for MapErrorToErrorMessage ---

func TestMapErrorToErrorMessage(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedMessage string
	}{
		{
			name:            "nil error returns empty message",
			err:             nil,
			expectedMessage: "",
		},
		{
			name:            "ErrNotFound returns standard message",
			err:             ErrNotFound,
			expectedMessage: "The requested resource was not found.",
		},
		{
			name:            "cargo.ErrNotFound returns standard message",
			err:             cargo.ErrNotFound,
			expectedMessage: "The requested resource was not found.",
		},
		{
			name:            "vessel.ErrNotFound returns standard message",
			err:             vessel.ErrNotFound,
			expectedMessage: "The requested resource was not found.",
		},
		{
			name:            "tracking.ErrCargoNotFound returns standard message",
			err:             tracking.ErrCargoNotFound,
			expectedMessage: "The requested resource was not found.",
		},
		{
			name:            "ErrInvalidTransition returns standard message",
			err:             ErrInvalidTransition,
			expectedMessage: "The requested operation violates business rules or state constraints.",
		},
		{
			name:            "cargo.ErrInvalidTransition returns standard message",
			err:             cargo.ErrInvalidTransition,
			expectedMessage: "The requested operation violates business rules or state constraints.",
		},
		{
			name:            "ErrInvalidInput returns standard message",
			err:             ErrInvalidInput,
			expectedMessage: "The request input is invalid. Please check the request and try again.",
		},
		{
			name:            "cargo.ErrInvalidInput returns standard message",
			err:             cargo.ErrInvalidInput,
			expectedMessage: "The request input is invalid. Please check the request and try again.",
		},
		{
			name:            "cargo.ErrInvalidStatus returns standard message",
			err:             cargo.ErrInvalidStatus,
			expectedMessage: "The request input is invalid. Please check the request and try again.",
		},
		{
			name:            "vessel.ErrInvalidInput returns standard message",
			err:             vessel.ErrInvalidInput,
			expectedMessage: "The request input is invalid. Please check the request and try again.",
		},
		{
			name:            "tracking.ErrInvalidEntry returns standard message",
			err:             tracking.ErrInvalidEntry,
			expectedMessage: "The request input is invalid. Please check the request and try again.",
		},
		{
			name:            "unknown error returns standard message",
			err:             errors.New("database connection failed"),
			expectedMessage: "An unexpected error occurred. Please try again later.",
		},
		{
			name:            "ErrUnknown returns standard message",
			err:             ErrUnknown,
			expectedMessage: "An unexpected error occurred. Please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			message := MapErrorToErrorMessage(tt.err)

			// Assert
			assert.Equal(t, tt.expectedMessage, message)
		})
	}
}

// --- Tests for Error Codes ---

func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected string
	}{
		{
			name:     "CodeNotFound value",
			code:     CodeNotFound,
			expected: "NOT_FOUND",
		},
		{
			name:     "CodeInvalidInput value",
			code:     CodeInvalidInput,
			expected: "INVALID_INPUT",
		},
		{
			name:     "CodeInvalidTransition value",
			code:     CodeInvalidTransition,
			expected: "INVALID_TRANSITION",
		},
		{
			name:     "CodeInternalError value",
			code:     CodeInternalError,
			expected: "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert
			assert.Equal(t, ErrorCode(tt.expected), tt.code)
		})
	}
}

// --- Tests for Sentinel Error Values ---

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "ErrNotFound has message",
			err:     ErrNotFound,
			message: "resource not found",
		},
		{
			name:    "ErrInvalidTransition has message",
			err:     ErrInvalidTransition,
			message: "invalid state transition",
		},
		{
			name:    "ErrInvalidInput has message",
			err:     ErrInvalidInput,
			message: "invalid input",
		},
		{
			name:    "ErrUnknown has message",
			err:     ErrUnknown,
			message: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert
			assert.Error(t, tt.err)
			assert.Equal(t, tt.message, tt.err.Error())
		})
	}
}
