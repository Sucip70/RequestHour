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

func randomHex(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
