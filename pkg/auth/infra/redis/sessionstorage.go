package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/auth"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/redis"
	"time"
)

type sessionJSONSchema struct {
	UserID uuid.UUID `json:"user_id"`
	Login  string    `json:"login"`
}

type sessionStorage struct {
	client redis.Client
}

func (s *sessionStorage) Add(sessionID string, session *auth.Session, ttl time.Duration) error {
	err := s.client.Set(s.getSessionKey(sessionID), s.encodeSession(session), &ttl)
	if err != nil {
		return fmt.Errorf("failed to add session: %w", err)
	}
	return nil
}

func (s *sessionStorage) Get(sessionID string) (*auth.Session, error) {
	data, err := s.client.Get(s.getSessionKey(sessionID))
	if errors.Is(err, redis.ErrKeyDoesNotExist) {
		return nil, auth.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	session, err := s.decodeSession(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	return session, err
}

func (s *sessionStorage) Remove(sessionID string) error {
	err := s.client.Del(s.getSessionKey(sessionID))
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return err
}

func (s *sessionStorage) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s:data", sessionID)
}

func (s *sessionStorage) encodeSession(session *auth.Session) string {
	bytes, _ := json.Marshal(sessionJSONSchema{
		UserID: session.UserID,
		Login:  session.Login,
	})
	return string(bytes)
}

func (s *sessionStorage) decodeSession(session string) (*auth.Session, error) {
	var result sessionJSONSchema
	err := json.Unmarshal([]byte(session), &result)
	if err != nil {
		return nil, err
	}

	return &auth.Session{
		UserID: result.UserID,
		Login:  result.Login,
	}, nil
}

func NewSessionStorage(client redis.Client) auth.SessionStorage {
	return &sessionStorage{client: client}
}
