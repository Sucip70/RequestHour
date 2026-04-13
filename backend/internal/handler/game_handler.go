package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"requesthour/backend/internal/model"
	"requesthour/backend/internal/service"
)

type GameHandler struct {
	svc *service.GameService
}

func NewGameHandler(svc *service.GameService) *GameHandler {
	return &GameHandler{svc: svc}
}

func sessionFromRequest(r *http.Request) string {
	if s := strings.TrimSpace(r.Header.Get("X-Session")); s != "" {
		return s
	}
	const p = "Bearer "
	if h := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(h), strings.ToLower(p)) {
		return strings.TrimSpace(h[len(p):])
	}
	return ""
}

// Question godoc
//
//	@Summary		Start song quiz round
//	@Description	Four random titles from gm_songs not yet in games; sets tr_session.current; returns shuffled titles and audioToken (AES-GCM: link+session+songId).
//	@Tags			game
//	@Produce		json
//	@Param			X-Session	header	string	true	"Session token"
//	@Success		200	{object}	model.GameQuestionResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		409	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/game/question [post]
func (h *GameHandler) Question(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromRequest(r)
	if sess == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "X-Session or Authorization Bearer header is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	out, err := h.svc.NextQuestion(ctx, sess)
	if err != nil {
		writeGameErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// Audio godoc
//
//	@Summary		Decrypt audio token
//	@Description	Returns the song link if the token matches this session and tr_session.current.
//	@Tags			game
//	@Accept			json
//	@Produce		json
//	@Param			X-Session	header	string					true	"Session token"
//	@Param			body		body	model.GameAudioRequest	true	"audioToken from POST /game/question"
//	@Success		200	{object}	model.GameAudioResponse
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/game/audio [post]
func (h *GameHandler) Audio(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromRequest(r)
	if sess == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "X-Session or Authorization Bearer header is required"})
		return
	}
	var req model.GameAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid JSON body"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	link, err := h.svc.RevealAudio(ctx, sess, req.AudioToken)
	if err != nil {
		writeGameErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model.GameAudioResponse{Link: link})
}

// Answer godoc
//
//	@Summary		Submit answer
//	@Description	Compares title to gm_songs for tr_session.current. Correct: append id to games. Wrong: reset games to empty.
//	@Tags			game
//	@Accept			json
//	@Produce		json
//	@Param			X-Session	header	string					true	"Session token"
//	@Param			body		body	model.GameAnswerRequest	true	"Chosen title"
//	@Success		200	{object}	model.GameAnswerResponse
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/game/answer [post]
func (h *GameHandler) Answer(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromRequest(r)
	if sess == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "X-Session or Authorization Bearer header is required"})
		return
	}
	var req model.GameAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid JSON body"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	out, err := h.svc.Answer(ctx, sess, req.Title)
	if err != nil {
		writeGameErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func writeGameErr(w http.ResponseWriter, err error) {
	log.Printf("game: %v", err)
	switch {
	case errors.Is(err, service.ErrGameSessionNotFound):
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "session not found"})
	case errors.Is(err, service.ErrNotEnoughSongs):
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: "not enough songs remaining for a new round"})
	case errors.Is(err, service.ErrNoActiveQuestion):
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "no active question; start a round with POST /game/question"})
	case errors.Is(err, service.ErrInvalidAudioToken):
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid or expired audio token"})
	default:
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "game operation failed"})
	}
}
