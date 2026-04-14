package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/jackc/pgx/v5"

	"requesthour/backend/internal/gamecrypto"
	"requesthour/backend/internal/model"
	"requesthour/backend/internal/repository"
	"requesthour/backend/internal/youtubeclip"
)

var (
	ErrGameSessionNotFound = errors.New("session not found")
	ErrNotEnoughSongs    = errors.New("not enough songs remaining")
	ErrNoActiveQuestion  = errors.New("no active question")
	ErrInvalidAudioToken = errors.New("invalid audio token")
	ErrAudioClipFailed   = errors.New("audio clip extraction failed")
)

type GameService struct {
	sessions *repository.SessionRepository
	songs    *repository.SongRepository
	key      []byte
	clip     *youtubeclip.Extractor
}

func NewGameService(sessions *repository.SessionRepository, songs *repository.SongRepository, gameSecret string, clip *youtubeclip.Extractor) *GameService {
	return &GameService{
		sessions: sessions,
		songs:    songs,
		key:      gamecrypto.Key(gameSecret),
		clip:     clip,
	}
}

type idTitle struct {
	id    int
	title string
	link  string
}

// NextQuestion picks 4 random songs not in games, sets current, returns shuffled titles and encrypted audio payload.
func (g *GameService) NextQuestion(ctx context.Context, session string) (*model.GameQuestionResponse, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return nil, ErrGameSessionNotFound
	}
	found, games, _, err := g.sessions.GetSessionState(ctx, session)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrGameSessionNotFound
	}

	ids, err := g.songs.RandomSongIDsExcluding(ctx, games, 4)
	if err != nil {
		return nil, err
	}
	if len(ids) < 4 {
		return nil, ErrNotEnoughSongs
	}

	correctIdx := rand.IntN(4)
	correctID := ids[correctIdx]

	pairs := make([]idTitle, 4)
	for i, id := range ids {
		title, link, err := g.songs.GetSongTitleLink(ctx, id)
		if err != nil {
			return nil, err
		}
		pairs[i] = idTitle{id: id, title: title, link: link}
	}

	correctLink := pairs[correctIdx].link

	rand.Shuffle(4, func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})
	titles := make([]string, 4)
	for i := range pairs {
		titles[i] = pairs[i].title
	}

	token, err := gamecrypto.EncryptPayload(g.key, gamecrypto.AudioPayload{
		Link:    correctLink,
		Session: session,
		SongID:  correctID,
	})
	if err != nil {
		return nil, err
	}

	if err := g.sessions.SetCurrent(ctx, session, correctID); err != nil {
		return nil, err
	}

	return &model.GameQuestionResponse{Titles: titles, AudioToken: token}, nil
}

// RevealAudio decrypts audioToken and returns the link only if it matches this session and current question.
func (g *GameService) RevealAudio(ctx context.Context, session, audioToken string) (string, error) {
	session = strings.TrimSpace(session)
	audioToken = strings.TrimSpace(audioToken)
	if session == "" || audioToken == "" {
		return "", ErrInvalidAudioToken
	}
	payload, err := gamecrypto.DecryptPayload(g.key, audioToken)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidAudioToken, err)
	}
	if payload.Session != session {
		return "", ErrInvalidAudioToken
	}
	found, _, current, err := g.sessions.GetSessionState(ctx, session)
	if err != nil {
		return "", err
	}
	if !found || current == nil || *current != payload.SongID {
		return "", ErrInvalidAudioToken
	}
	return payload.Link, nil
}

// AudioClipMP3 verifies the token like RevealAudio, then returns the first 10 seconds of the YouTube track as MP3 bytes (requires yt-dlp + ffmpeg).
func (g *GameService) AudioClipMP3(ctx context.Context, session, audioToken string) ([]byte, error) {
	link, err := g.RevealAudio(ctx, session, audioToken)
	if err != nil {
		return nil, err
	}
	data, err := g.clip.FirstSecondsMP3(ctx, link, 10)
	if err != nil {
		if errors.Is(err, youtubeclip.ErrNotYouTube) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", ErrAudioClipFailed, err)
	}
	return data, nil
}

// Answer compares the submitted title to the song in current; updates games or resets on wrong answer.
func (g *GameService) Answer(ctx context.Context, session, answerTitle string) (*model.GameAnswerResponse, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return nil, ErrGameSessionNotFound
	}
	found, games, current, err := g.sessions.GetSessionState(ctx, session)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrGameSessionNotFound
	}
	if current == nil {
		return nil, ErrNoActiveQuestion
	}

	dbTitle, _, err := g.songs.GetSongTitleLink(ctx, *current)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActiveQuestion
		}
		return nil, err
	}

	norm := func(s string) string { return strings.TrimSpace(s) }
	if norm(answerTitle) != norm(dbTitle) {
		if err := g.sessions.ResetGames(ctx, session); err != nil {
			return nil, err
		}
		return &model.GameAnswerResponse{Correct: false, Games: []int{}}, nil
	}

	if err := g.sessions.AppendGameClearCurrent(ctx, session, *current); err != nil {
		return nil, err
	}
	updatedGames := append(append([]int(nil), games...), *current)
	return &model.GameAnswerResponse{Correct: true, Games: updatedGames}, nil
}
