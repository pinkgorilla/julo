package auth

import (
	"context"
	"julo/internal/account"
	"sync"
)

type Session struct {
	Token   string
	Account account.Account
}

type SessionManager interface {
	StoreSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
}

type InMemorySessionManager struct {
	store sync.Map
}

func NewInMemorySessionManager() SessionManager {
	return &InMemorySessionManager{
		store: sync.Map{},
	}
}

func (m *InMemorySessionManager) StoreSession(ctx context.Context, session Session) error {
	m.store.Store(session.Token, &session)
	return nil
}

func (m *InMemorySessionManager) GetSession(ctx context.Context, token string) (*Session, error) {
	v, ok := m.store.Load(token)
	if !ok {
		return nil, ErrSessionNotFound
	}

	return v.(*Session), nil
}

var sessionManager SessionManager
var once sync.Once

func init() {
	once.Do(func() {
		if sessionManager == nil {
			sessionManager = NewInMemorySessionManager()
		}
	})
}

func StoreSession(ctx context.Context, session Session) error {
	return sessionManager.StoreSession(ctx, session)
}

func GetSession(ctx context.Context, token string) (*Session, error) {
	return sessionManager.GetSession(ctx, token)
}

type key string

const (
	sessionKey = key("session-key")
)

func SessionIntoContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func SessionFromContext(ctx context.Context) *Session {
	v := ctx.Value(sessionKey)
	session, ok := v.(*Session)
	if !ok {
		return nil
	}

	return session
}
