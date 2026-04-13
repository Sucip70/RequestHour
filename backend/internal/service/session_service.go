package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"requesthour/backend/internal/model"
	"requesthour/backend/internal/repository"
)

var (
	ErrGenerateToken = errors.New("generate session token")
	ErrSaveSession   = errors.New("save session")
	ErrCheckSession  = errors.New("check session")
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

// CreateSession generates a random token and persists it.
func (s *SessionService) CreateSession(ctx context.Context) (*model.Session, error) {
	token, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGenerateToken, err)
	}
	if err := s.repo.InsertSession(ctx, token); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSaveSession, err)
	}
	return &model.Session{Session: token}, nil
}

// LookupSession returns whether the token exists, games, and current question id.
func (s *SessionService) LookupSession(ctx context.Context, session string) (model.SessionExistsResponse, error) {
	if session == "" {
		return model.SessionExistsResponse{Exists: false, Games: []int{}}, nil
	}
	found, games, current, err := s.repo.GetSessionState(ctx, session)
	if err != nil {
		return model.SessionExistsResponse{}, fmt.Errorf("%w: %w", ErrCheckSession, err)
	}
	if games == nil {
		games = []int{}
	}
	return model.SessionExistsResponse{Exists: found, Games: games, Current: current}, nil
}

func randomHex(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
