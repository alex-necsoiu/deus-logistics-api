package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateRequiredString tests the ValidateRequiredString validator.
func TestValidateRequiredString(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid non-empty string",
			fieldName: "Name",
			value:     "test cargo",
			wantErr:   false,
		},
		{
			name:      "empty string",
			fieldName: "Name",
			value:     "",
			wantErr:   true,
			errMsg:    "Name cannot be empty",
		},
		{
			name:      "whitespace only",
			fieldName: "Description",
			value:     "   ",
			wantErr:   true,
			errMsg:    "Description cannot be empty",
		},
		{
			name:      "string with leading/trailing whitespace",
			fieldName: "Code",
			value:     "  ABC123  ",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredString(tt.fieldName, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePositiveFloat tests the ValidatePositiveFloat validator.
func TestValidatePositiveFloat(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     float64
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "positive float",
			fieldName: "Weight",
			value:     100.5,
			wantErr:   false,
		},
		{
			name:      "zero value",
			fieldName: "Weight",
			value:     0,
			wantErr:   true,
			errMsg:    "Weight must be greater than 0",
		},
		{
			name:      "negative value",
			fieldName: "Price",
			value:     -50.0,
			wantErr:   true,
			errMsg:    "Price must be greater than 0",
		},
		{
			name:      "small positive float",
			fieldName: "Quantity",
			value:     0.001,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveFloat(tt.fieldName, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateFloatRange tests the ValidateFloatRange validator.
func TestValidateFloatRange(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     float64
		min       float64
		max       float64
		wantErr   bool
	}{
		{
			name:      "value within range",
			fieldName: "Temperature",
			value:     20.0,
			min:       -10.0,
			max:       50.0,
			wantErr:   false,
		},
		{
			name:      "value at min boundary",
			fieldName: "Humidity",
			value:     0,
			min:       0,
			max:       100,
			wantErr:   false,
		},
		{
			name:      "value at max boundary",
			fieldName: "Humidity",
			value:     100,
			min:       0,
			max:       100,
			wantErr:   false,
		},
		{
			name:      "value below min",
			fieldName: "Percentage",
			value:     -10.0,
			min:       0,
			max:       100,
			wantErr:   true,
		},
		{
			name:      "value above max",
			fieldName: "Percentage",
			value:     150.0,
			min:       0,
			max:       100,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFloatRange(tt.fieldName, tt.value, tt.min, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateMaxFloat tests the ValidateMaxFloat validator.
func TestValidateMaxFloat(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     float64
		max       float64
		wantErr   bool
	}{
		{
			name:      "value below max",
			fieldName: "Discount",
			value:     20.0,
			max:       50.0,
			wantErr:   false,
		},
		{
			name:      "value at max",
			fieldName: "Fee",
			value:     100.0,
			max:       100.0,
			wantErr:   false,
		},
		{
			name:      "value exceeds max",
			fieldName: "Surcharge",
			value:     150.0,
			max:       100.0,
			wantErr:   true,
		},
		{
			name:      "negative value below max",
			fieldName: "Adjustment",
			value:     -50.0,
			max:       0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxFloat(tt.fieldName, tt.value, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateCargoStatus tests the ValidateCargoStatus validator.
func TestValidateCargoStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{
			name:    "valid status pending",
			status:  "pending",
			wantErr: false,
		},
		{
			name:    "valid status in_transit",
			status:  "in_transit",
			wantErr: false,
		},
		{
			name:    "valid status delivered",
			status:  "delivered",
			wantErr: false,
		},
		{
			name:    "invalid status",
			status:  "cancelled",
			wantErr: true,
		},
		{
			name:    "empty status",
			status:  "",
			wantErr: true,
		},
		{
			name:    "uppercase status",
			status:  "PENDING",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCargoStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateVesselCapacity tests the ValidateVesselCapacity validator.
func TestValidateVesselCapacity(t *testing.T) {
	tests := []struct {
		name     string
		capacity float64
		wantErr  bool
	}{
		{
			name:     "valid capacity in range",
			capacity: 500.0,
			wantErr:  false,
		},
		{
			name:     "minimum valid capacity",
			capacity: 0.1,
			wantErr:  false,
		},
		{
			name:     "maximum valid capacity",
			capacity: 1000.0,
			wantErr:  false,
		},
		{
			name:     "zero capacity",
			capacity: 0,
			wantErr:  true,
		},
		{
			name:     "negative capacity",
			capacity: -100.0,
			wantErr:  true,
		},
		{
			name:     "exceeds maximum",
			capacity: 1001.0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVesselCapacity(tt.capacity)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidationErrors tests the ValidationErrors collection.
func TestValidationErrors(t *testing.T) {
	t.Run("empty errors is valid", func(t *testing.T) {
		ve := &ValidationErrors{}
		assert.True(t, ve.Valid())
		assert.Equal(t, 0, ve.Count())
		assert.Equal(t, "", ve.Error())
	})

	t.Run("single error", func(t *testing.T) {
		ve := &ValidationErrors{}
		ve.Add("field required")
		assert.False(t, ve.Valid())
		assert.Equal(t, 1, ve.Count())
		assert.Equal(t, "field required", ve.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		ve := &ValidationErrors{}
		ve.Add("name is required")
		ve.Add("email is invalid")
		ve.Add("age must be positive")
		assert.False(t, ve.Valid())
		assert.Equal(t, 3, ve.Count())
		errMsg := ve.Error()
		assert.Contains(t, errMsg, "validation failed with 3 errors")
		assert.Contains(t, errMsg, "name is required")
		assert.Contains(t, errMsg, "email is invalid")
		assert.Contains(t, errMsg, "age must be positive")
	})

	t.Run("ignore empty error messages", func(t *testing.T) {
		ve := &ValidationErrors{}
		ve.Add("error 1")
		ve.Add("")
		ve.Add("error 2")
		assert.Equal(t, 2, ve.Count())
	})

	t.Run("AddError with error", func(t *testing.T) {
		ve := &ValidationErrors{}
		ve.AddError(ValidateRequiredString("name", ""))
		assert.Equal(t, 1, ve.Count())
		assert.False(t, ve.Valid())
	})

	t.Run("AddError with nil", func(t *testing.T) {
		ve := &ValidationErrors{}
		ve.AddError(nil)
		assert.Equal(t, 0, ve.Count())
		assert.True(t, ve.Valid())
	})
}
