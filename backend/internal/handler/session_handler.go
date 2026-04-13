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
//	@Description	Generates a random token and stores it in tr_session.session.
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

// CheckSession godoc
//
//	@Summary		Check session exists
//	@Description	Returns whether the session exists and the tr_session.games integer array (empty when missing).
//	@Tags			session
//	@Produce		json
//	@Param			session	path	string	true	"Session token (hex)"
//	@Success		200	{object}	model.SessionExistsResponse
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/session/{session} [get]
func (h *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("session")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "session is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	out, err := h.svc.LookupSession(ctx, token)
	if err != nil {
		log.Printf("check session: %v", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to check session"})
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
