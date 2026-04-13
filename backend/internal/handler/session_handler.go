package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"requesthour/backend/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	sess, err := h.svc.CreateSession(ctx)
	if err != nil {
		log.Printf("create session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		msg := "failed to save session"
		if errors.Is(err, service.ErrGenerateToken) {
			msg = "failed to generate session"
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sess)
}
