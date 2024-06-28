package api

// Error is an error from the REST API
type Error struct {
	Message     string `json:"message"`       // Message is the error message
	ErrorCode   string `json:"error_code"`    // ErrorCode is the string name of the error
	VmErrorCode uint64 `json:"vm_error_code"` // VmErrorCode is the number of the failure, optional 0 if not set
}
