# HTTP Request Validation

## Overview

This document describes the strict request validation system implemented at the HTTP layer of the DEUS Logistics API. All incoming requests are validated BEFORE any use case execution to ensure data integrity and provide clear, consistent error responses.

## Validation Strategy

### Layer-by-Layer Validation

Validation occurs in multiple layers to ensure defense-in-depth:

1. **JSON Binding Layer** (`ShouldBindJSON`)
   - Gin framework validates struct tags: `binding:"required"`
   - Enforces required fields and basic type constraints
   - Rejects invalid request payloads with HTTP 400

2. **DTO Validation Layer** (`Validate()` methods)
   - Business logic validation of individual fields
   - Status values, numeric ranges, empty strings after trim
   - Called explicitly in each handler BEFORE use case execution

3. **Centralized Validators** (`internal/validation`)
   - Reusable validation functions for consistency
   - Prevents duplicate validation logic
   - Easy to extend with new validators

### Validation Timing

```
HTTP Request
    ↓
Content-Type Check (middleware)
    ↓
JSON Binding (ShouldBindJSON)
    ↓
DTO Validation (req.Validate())
    ↓
Use Case Execution
    ↓
HTTP Response
```

**Critical:** Validation happens BEFORE use case execution. Services never receive invalid data.

## Validation Components

### 1. Centralized Validators (`internal/validation/validators.go`)

Reusable validation functions:

- `ValidateRequiredString(fieldName, value)` - Non-empty after trim
- `ValidatePositiveFloat(fieldName, value)` - Value > 0
- `ValidateFloatRange(fieldName, value, min, max)` - Value in [min, max]
- `ValidateMaxFloat(fieldName, value, max)` - Value ≤ max
- `ValidateCargoStatus(status)` - Valid cargo status
- `ValidateVesselCapacity(capacity)` - Valid vessel capacity (1-1000 tons)

### 2. Request DTOs (`internal/transport/http/dto.go`)

Each request DTO has a `Validate()` method:

```go
type CreateCargoRequest struct {
    Name        string    `json:"name" binding:"required"`
    Description string    `json:"description"`
    Weight      float64   `json:"weight" binding:"required,gt=0"`
    VesselID    uuid.UUID `json:"vessel_id" binding:"required"`
}

func (r *CreateCargoRequest) Validate() error {
    if err := validation.ValidateRequiredString("name", r.Name); err != nil {
        return err
    }
    if err := validation.ValidatePositiveFloat("weight", r.Weight); err != nil {
        return err
    }
    if r.VesselID == uuid.Nil {
        return errors.New("vessel_id is required and must be a valid UUID")
    }
    return nil
}
```

### 3. Validation Middleware (`internal/transport/middleware/validation.go`)

- `ValidateJSONContentType()` - Ensures POST/PUT/PATCH requests have `Content-Type: application/json`

## Handled Request Types

### Cargo Operations

#### CreateCargoRequest
- **Constraints:**
  - `name`: Required, non-empty after trim
  - `description`: Optional, can be empty
  - `weight`: Required, must be > 0
  - `vessel_id`: Required, must be valid UUID

- **Validation:**
  ```
  POST /api/v1/cargoes
  {
    "name": "Electronics Shipment",
    "description": "High-value electronics",
    "weight": 150.5,
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000"
  }
  ```

- **Invalid Examples:**
  - Empty name: `{"name": "", "weight": 100, "vessel_id": "..."}`  → 400, "name cannot be empty"
  - Negative weight: `{"name": "Cargo", "weight": -5, "vessel_id": "..."}`  → 400, "weight must be greater than 0"
  - Invalid UUID: `{"name": "Cargo", "weight": 100, "vessel_id": "invalid"}`  → 400, invalid request body

#### UpdateCargoStatusRequest
- **Constraints:**
  - `status`: Required, must be one of: `pending`, `in_transit`, `delivered`

- **Invalid Examples:**
  - Invalid status: `{"status": "shipped"}` → 400, "invalid status: shipped, must be one of: pending, in_transit, delivered"
  - Missing status: `{}` → 400, invalid request body

### Vessel Operations

#### CreateVesselRequest
- **Constraints:**
  - `name`: Required, non-empty after trim
  - `capacity`: Required, must be > 0 and ≤ 1000 tons
  - `current_location`: Required, non-empty after trim

- **Invalid Examples:**
  - Capacity too high: `{"name": "Ship", "capacity": 1500, "current_location": "..."}` → 400, "capacity cannot exceed 1000 tons"
  - Negative capacity: `{"name": "Ship", "capacity": -100, "current_location": "..."}` → 400, "capacity must be greater than 0"
  - Empty location: `{"name": "Ship", "capacity": 500, "current_location": ""}` → 400, "current_location cannot be empty"

#### UpdateVesselLocationRequest
- **Constraints:**
  - `current_location`: Required, non-empty after trim

- **Invalid Examples:**
  - Empty location: `{"current_location": ""}` → 400, "current_location cannot be empty"
  - Whitespace only: `{"current_location": "   "}` → 400, "current_location cannot be empty"

### Tracking Operations

#### AddTrackingRequest
- **Constraints:**
  - `location`: Required, non-empty after trim
  - `status`: Required, one of: `pending`, `in_transit`, `delivered`
  - `note`: Optional

- **Invalid Examples:**
  - Empty location: `{"location": "", "status": "in_transit"}` → 400, "location cannot be empty"
  - Invalid status: `{"location": "Port A", "status": "unknown"}` → 400, "invalid status"

## Error Responses

All validation errors return HTTP 400 with standardized format:

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "field cannot be empty",
    "request_id": "req-123456"
  }
}
```

## Handler Pattern

All handlers follow this validation pattern:

```go
func (h *Handler) SomeEndpoint(c *gin.Context) {
    ctx := c.Request.Context()
    logger := zerolog.Ctx(ctx)

    // 1. Parse request body
    var req SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Error().Err(err).Msg("invalid request body")
        c.JSON(http.StatusBadRequest, response.ErrorResponse{
            Error: response.ErrorDetail{
                Code:      response.CodeInvalidInput,
                Message:   response.MsgInvalidRequestBody,
                RequestID: c.GetString(response.CtxRequestID),
            },
        })
        return
    }

    // 2. Validate business rules BEFORE use case execution
    if err := req.Validate(); err != nil {
        logger.Error().Err(err).Msg("validation failed")
        c.JSON(http.StatusBadRequest, response.ErrorResponse{
            Error: response.ErrorDetail{
                Code:      response.CodeInvalidInput,
                Message:   err.Error(),
                RequestID: c.GetString(response.CtxRequestID),
            },
        })
        return
    }

    // 3. Execute use case with validated data
    result, err := h.app.SomeUseCase.Execute(ctx, convertToInput(req))
    if err != nil {
        // Handle use case errors
        status := mapErrorToStatus(err)
        c.JSON(status, response.ErrorResponse{...})
        return
    }

    // 4. Return success
    c.JSON(http.StatusOK, response.SuccessResponse{Data: result})
}
```

## Extending Validation

To add new validators:

1. **Add to `internal/validation/validators.go`:**
   ```go
   func ValidateMyField(fieldName string, value string) error {
       // Implement validation logic
       return nil // or error
   }
   ```

2. **Use in DTO `Validate()` method:**
   ```go
   func (r *MyRequest) Validate() error {
       if err := validation.ValidateMyField("field", r.Field); err != nil {
           return err
       }
       return nil
   }
   ```

## Constraints Enforced

### Cargo
- ✅ Name: non-empty
- ✅ Weight: positive number
- ✅ VesselID: valid UUID
- ✅ Status: pending | in_transit | delivered
- ✅ Status transitions: pending → in_transit → delivered (enforced by domain)

### Vessel
- ✅ Name: non-empty
- ✅ Capacity: 0 < capacity ≤ 1000 tons
- ✅ Location: non-empty

### Tracking
- ✅ Location: non-empty
- ✅ Status: pending | in_transit | delivered
- ✅ Note: optional

## Testing Validation

All validation errors return HTTP 400 with clear messages. See `internal/transport/http/*_handler_test.go` for validation test cases.

## Key Design Principles

1. **Fail Fast:** Validation happens at HTTP layer, before business logic
2. **Clear Errors:** Users receive specific, actionable error messages
3. **No Leakage:** Database/system errors never reach clients
4. **Consistency:** All handlers follow the same validation pattern
5. **Centralization:** Reusable validators prevent duplication
6. **Type Safety:** Go struct tags + explicit validation methods
