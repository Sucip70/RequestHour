package model

// ErrorResponse is the JSON body for API errors.
type ErrorResponse struct {
	Error string `json:"error" example:"failed to save session"`
}
