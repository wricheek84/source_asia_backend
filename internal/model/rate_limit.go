package model

import "encoding/json"

// RequestInput represents the incoming payload for POST /request.
type RequestInput struct {
	UserID  string          `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

// UserStats represents the metrics returned by GET /stats per user.
type UserStats struct {
	AcceptedCurrentWindow int `json:"accepted_current_window"`
	RejectedCumulative    int `json:"rejected_cumulative"`
}