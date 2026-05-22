package model

import "encoding/json"

type RequestInput struct {
	UserID  string          `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}


type UserStats struct {
	AcceptedCurrentWindow int `json:"accepted_current_window"`
	RejectedCumulative    int `json:"rejected_cumulative"`
}