package model

// Session is a persisted client session token returned to the API consumer.
type Session struct {
	Session string `json:"session"`
}
