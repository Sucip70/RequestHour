package model

// SessionExistsResponse is returned by GET /session/{session}.
type SessionExistsResponse struct {
	Exists  bool  `json:"exists" example:"true"`
	Games   []int `json:"games" swaggertype:"array,integer"`
	Current *int  `json:"current,omitempty" swaggertype:"integer"`
}
