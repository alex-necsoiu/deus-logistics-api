package response

// SuccessResponse is the standard JSON response for successful requests.
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// ErrorResponse is the standard JSON response for error requests.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Meta contains response metadata.
type Meta struct {
	RequestID string `json:"request_id"`
}
