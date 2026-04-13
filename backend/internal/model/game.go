package model

// GameQuestionResponse is returned when a new quiz round starts.
type GameQuestionResponse struct {
	Titles     []string `json:"titles" swaggertype:"array,string"`
	AudioToken string   `json:"audioToken"`
}

// GameAnswerRequest is the body for POST /game/answer.
type GameAnswerRequest struct {
	Title string `json:"title" example:"Song Title"`
}

// GameAnswerResponse reports result and updated games progress.
type GameAnswerResponse struct {
	Correct bool `json:"correct" example:"true"`
	Games   []int `json:"games" swaggertype:"array,integer"`
}

// GameAudioRequest decrypts the audio token for playback.
type GameAudioRequest struct {
	AudioToken string `json:"audioToken"`
}

// GameAudioResponse exposes the link after server-side verification.
type GameAudioResponse struct {
	Link string `json:"link"`
}
