package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/rzajac/zltest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sdk/proto/a23n/v1"
	"github.com/ashep/a23n/server/credentials"
	"github.com/ashep/a23n/server/handler"
)

type AuthenticateTestSuite struct {
	suite.Suite

	api     *api.APIMock
	handler *handler.Handler
	logger  *zltest.Tester
}

func (s *AuthenticateTestSuite) SetupTest() {
	lt := zltest.New(s.T())
	l := lt.Logger().Level(zerolog.DebugLevel)

	s.api = &api.APIMock{}
	s.handler = handler.New(s.api, time.Second*5, time.Second*10, l)
	s.logger = lt
}

func (s *AuthenticateTestSuite) TearDownTest() {
	s.api.AssertExpectations(s.T())
}

func (s *AuthenticateTestSuite) TestNoAuthorizationHeader() {
	req := connect.NewRequest(&v1.AuthenticateRequest{})
	_, err := s.handler.Authenticate(context.Background(), req)
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))
	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestRequestEmptyEntityId() {
	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeInvalidArgument, errors.New("empty entity id")))
	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestEntityNotFound() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{}, api.ErrNotFound)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"warn","entity_id":"anEntityID","message":"entity not found"}`, l.String())
}

func (s *AuthenticateTestSuite) TestAPIGetEntityError() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{}, errors.New("theGetEntityError"))

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"theGetEntityError","entity_id":"anEntityID","message":"failed to get entity"}`, l.String())
}

func (s *AuthenticateTestSuite) TestAPICheckSecretError() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{}, nil)

	s.api.
		On("CheckSecret", "anEntityID", "aPassword").
		Return(false, errors.New("theCheckSecretError"))

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"theCheckSecretError","entity_id":"anEntityID","message":"check secret failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestInvalidSecret() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{}, nil)

	s.api.
		On("CheckSecret", "anEntityID", "aPassword").
		Return(false, nil)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestOutOfScope() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{ID: "anEntityID"}, nil)

	s.api.
		On("CheckSecret", "anEntityID", "aPassword").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"aScopeItem"}).
		Return(false)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"aScopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodePermissionDenied, nil))

	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestGetAccessTokenExpirationTimeError() {
	cl := &api.ClaimsMock{}
	cl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, errors.New("theGetExpirationTimeError"))

	tk := &api.TokenMock{}
	tk.
		On("Claims").
		Return(cl)

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{ID: "anEntityID"}, nil)

	s.api.
		On("CheckSecret", "anEntityID", "aPassword").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"aScopeItem"}).
		Return(true)

	s.api.
		On("CreateToken", "anEntityID", []string(nil), time.Second*5).
		Return(tk)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"aScopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"theGetExpirationTimeError","entity_id":"anEntityID","message":"get access token expiration time failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestGetAccessTokenSignedStringError() {
	cl := &api.ClaimsMock{}
	cl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, nil)

	tk := &api.TokenMock{}
	tk.
		On("Claims").
		Return(cl)
	tk.
		On("SignedString", "aSecretKey").
		Return("", errors.New("theSignedStringError"))

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "anEntityID").
		Return(api.Entity{ID: "anEntityID"}, nil)

	s.api.
		On("CheckSecret", "anEntityID", "aPassword").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"aScopeItem"}).
		Return(true)

	s.api.
		On("CreateToken", "anEntityID", []string(nil), time.Second*5).
		Return(tk)

	s.api.
		On("SecretKey").
		Return("aSecretKey")

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "anEntityID",
		Password: "aPassword",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"aScopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"theSignedStringError","entity_id":"anEntityID","message":"get access token signed string failed"}`, l.String())
}

func TestHandler_Authenticate(t *testing.T) {
	suite.Run(t, new(AuthenticateTestSuite))
}
