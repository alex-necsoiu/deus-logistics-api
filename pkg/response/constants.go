package response

// Error codes returned in API responses.
const (
	CodeInvalidInput     = "INVALID_INPUT"
	CodeInvalidStatus    = "INVALID_STATUS"
	CodeInvalidEntry     = "INVALID_ENTRY"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeCargoNotFound    = "CARGO_NOT_FOUND"
	CodeVesselNotFound   = "VESSEL_NOT_FOUND"
	CodeCapacityExceeded = "CAPACITY_EXCEEDED"
)

// Error messages returned in API responses.
const (
	MsgInvalidRequestBody     = "invalid request body"
	MsgInvalidCargoID         = "invalid cargo id"
	MsgInvalidVesselID        = "invalid vessel id"
	MsgFailedListCargoes      = "failed to list cargoes"
	MsgFailedListVessels      = "failed to list vessels"
	MsgFailedTrackingHistory  = "failed to get tracking history"
	MsgInternalServerError    = "internal server error"
)

// HTTP header and context keys.
const (
	HeaderRequestID = "X-Request-ID"
	CtxRequestID    = "request_id"
)
