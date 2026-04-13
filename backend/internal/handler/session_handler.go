package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"requesthour/backend/internal/model"
	"requesthour/backend/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// CreateSession godoc
//
//	@Summary		Create session
//	@Description	Generates a random token and stores it in gm_session.session.
//	@Tags			session
//	@Produce		json
//	@Success		201	{object}	model.Session
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/session [post]
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	sess, err := h.svc.CreateSession(ctx)
	if err != nil {
		log.Printf("create session: %v", err)
		msg := "failed to save session"
		if errors.Is(err, service.ErrGenerateToken) {
			msg = "failed to generate session"
		}
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: msg})
		return
	}

	writeJSON(w, http.StatusCreated, sess)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
