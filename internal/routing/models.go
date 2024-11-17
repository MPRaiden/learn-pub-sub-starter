package routing

import "time"

type PlayingState struct {
	IsPaused bool
}

type GameLog struct {
	CurrentTime time.Time `json:"current_time"`
	Message     string    `json:"message"`
	Username    string    `json:"username"`
}
