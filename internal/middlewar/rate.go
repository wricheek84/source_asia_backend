package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/wricheek84/source_asia_backend/internal/model"
	"github.com/wricheek84/source_asia_backend/internal/store"
)

// RateLimitMiddleware intercepts requests to validate the rolling window limit.
func RateLimitMiddleware(rlStore *store.RateLimitStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Only enforce rate limiting on the POST /request endpoint
			if r.Method != http.MethodPost || r.URL.Path != "/request" {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Read the body text safely without destroying it
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Failed to read request body"})
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 3. Extract just the user_id from the incoming payload
			var input model.RequestInput
			if err := json.Unmarshal(bodyBytes, &input); err != nil || input.UserID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Invalid input: user_id is required"})
				return
			}

			// 4. Check the clock and lock limits in our storage fridge
			allowed, _ := rlStore.IncrementAndCheck(input.UserID)
			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Too Many Requests"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}