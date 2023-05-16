package api

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

type TokenMock struct {
	mock.Mock
}

func (m *TokenMock) Claims() jwt.Claims {
	args := m.Called()
	return args.Get(0).(jwt.Claims)
}

func (m *TokenMock) SignedString(key interface{}) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

type APIMock struct {
	mock.Mock
}

func (m *APIMock) CheckSecret(hashed, secret string) (bool, error) {
	args := m.Called(hashed, secret)
	return args.Bool(0), args.Error(1)
}

func (m *APIMock) CreateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error {
	args := m.Called(ctx, id, secret, scope, attrs)
	return args.Error(0)
}

func (m *APIMock) UpdateEntity(ctx context.Context, id string, secret []byte, scope Scope, attrs Attrs) error {
	args := m.Called(ctx, id, secret, scope, attrs)
	return args.Error(0)
}

func (m *APIMock) GetEntity(ctx context.Context, id string) (Entity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *APIMock) CheckScope(target Scope, required Scope) bool {
	args := m.Called(target, required)
	return args.Bool(0)
}

func (m *APIMock) CreateToken(subject string, scope []string, ttl time.Duration) Token {
	args := m.Called(subject, scope, ttl)
	return args.Get(0).(Token)
}

func (m *APIMock) GetTokenSignedString(t Token) (string, error) {
	args := m.Called(t)
	return args.String(0), args.Error(1)
}

func (m *APIMock) ParseToken(t string) (TokenClaims, error) {
	args := m.Called(t)
	return args.Get(0).(TokenClaims), args.Error(1)
}
