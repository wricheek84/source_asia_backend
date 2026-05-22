package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/wricheek84/source_asia_backend/internal/model"
	"github.com/wricheek84/source_asia_backend/internal/store"
)


func RateLimitMiddleware(rlStore *store.RateLimitStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			if r.Method != http.MethodPost || r.URL.Path != "/request" {
				next.ServeHTTP(w, r)
				return
			}

			
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Failed to read request body"})
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		
			var input model.RequestInput
			if err := json.Unmarshal(bodyBytes, &input); err != nil || input.UserID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Invalid input: user_id is required"})
				return
			}

			
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