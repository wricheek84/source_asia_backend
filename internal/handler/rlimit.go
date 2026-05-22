package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wricheek84/source_asia_backend/internal/model"
	"github.com/wricheek84/source_asia_backend/internal/store"
)


type RateLimitHandler struct {
	rlStore *store.RateLimitStore
}


func NewRateLimitHandler(rlStore *store.RateLimitStore) *RateLimitHandler {
	return &RateLimitHandler{rlStore: rlStore}
}


func (h *RateLimitHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}


func (h *RateLimitHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Method not allowed"})
		return
	}

	
	userID := r.URL.Query().Get("user_id")
	if strings.TrimSpace(userID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Missing required query parameter: user_id"})
		return
	}

	stats := h.rlStore.GetStats(userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}