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
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeInvalidArgument, errors.New("empty entity id")))
	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestEntityNotFound() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{}, api.ErrNotFound)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"warn","entity_id":"entityID","message":"entity not found"}`, l.String())
}

func (s *AuthenticateTestSuite) TestAPIGetEntityError() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{}, errors.New("getEntityError"))

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"getEntityError","entity_id":"entityID","message":"failed to get entity"}`, l.String())
}

func (s *AuthenticateTestSuite) TestAPICheckSecretError() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(false, errors.New("theCheckSecretError"))

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"theCheckSecretError","entity_id":"entityID","message":"check secret failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestInvalidSecret() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(false, nil)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{}))
	s.Require().Equal(err, connect.NewError(connect.CodeUnauthenticated, nil))

	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestOutOfScope() {
	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(false)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodePermissionDenied, nil))

	s.Require().Nil(s.logger.LastEntry())
}

func (s *AuthenticateTestSuite) TestGetAccessTokenExpirationTimeError() {
	cl := &api.ClaimsMock{}
	cl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, errors.New("accessTokenExpirationTimeError"))

	tk := &api.TokenMock{}
	tk.
		On("Claims").
		Return(cl)

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(true)

	s.api.
		On("CreateToken", "entityID", []string(nil), time.Second*5).
		Return(tk)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"accessTokenExpirationTimeError","entity_id":"entityID","message":"get access token expiration time failed"}`, l.String())
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
		On("SignedString", "secretKey").
		Return("", errors.New("accessTokenSignedStringError"))

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(true)

	s.api.
		On("CreateToken", "entityID", []string(nil), time.Second*5).
		Return(tk)

	s.api.
		On("SecretKey").
		Return("secretKey")

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"accessTokenSignedStringError","entity_id":"entityID","message":"get access token signed string failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestGetRefreshTokenExpirationTimeError() {
	atCl := &api.ClaimsMock{}
	atCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, nil)

	at := &api.TokenMock{}
	at.
		On("Claims").
		Return(atCl)
	at.
		On("SignedString", "secretKey").
		Return("accessTokenSignedString", nil)

	rtCl := &api.ClaimsMock{}
	rtCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, errors.New("refreshTokenExpirationTimeError"))

	rt := &api.TokenMock{}
	rt.
		On("Claims").
		Return(rtCl)

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(true)

	s.api.
		On("SecretKey").
		Return("secretKey")

	s.api.
		On("CreateToken", "entityID", []string(nil), time.Second*5).
		Return(at)

	s.api.
		On("CreateToken", "entityID_refresh", []string(nil), time.Second*10).
		Return(rt)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"refreshTokenExpirationTimeError","entity_id":"entityID","message":"get refresh token expiration time failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestGetRefreshTokenSignedStringError() {
	atCl := &api.ClaimsMock{}
	atCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, nil)

	at := &api.TokenMock{}
	at.
		On("Claims").
		Return(atCl)
	at.
		On("SignedString", "secretKey").
		Return("accessTokenSignedString", nil)

	rtCl := &api.ClaimsMock{}
	rtCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{}, nil)

	rt := &api.TokenMock{}
	rt.
		On("Claims").
		Return(rtCl)
	rt.
		On("SignedString", "secretKey").
		Return("", errors.New("refreshTokenSignedStringError"))

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(true)

	s.api.
		On("SecretKey").
		Return("secretKey")

	s.api.
		On("CreateToken", "entityID", []string(nil), time.Second*5).
		Return(at)

	s.api.
		On("CreateToken", "entityID_refresh", []string(nil), time.Second*10).
		Return(rt)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	_, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().Equal(err, connect.NewError(connect.CodeInternal, nil))

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"error","error":"refreshTokenSignedStringError","entity_id":"entityID","message":"get refresh token signed string failed"}`, l.String())
}

func (s *AuthenticateTestSuite) TestOK() {
	atCl := &api.ClaimsMock{}
	atCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{Time: time.Unix(123456789, 0)}, nil)

	at := &api.TokenMock{}
	at.
		On("Claims").
		Return(atCl)
	at.
		On("SignedString", "secretKey").
		Return("accessTokenSignedString", nil)

	rtCl := &api.ClaimsMock{}
	rtCl.On("GetExpirationTime").
		Return(&jwt.NumericDate{Time: time.Unix(234567890, 0)}, nil)

	rt := &api.TokenMock{}
	rt.
		On("Claims").
		Return(rtCl)
	rt.
		On("SignedString", "secretKey").
		Return("refreshTokenSignedString", nil)

	s.api.
		On("GetEntity", mock.AnythingOfType("*context.valueCtx"), "entityID").
		Return(api.Entity{ID: "entityID"}, nil)

	s.api.
		On("CheckSecret", "entityID", "password").
		Return(true, nil)

	s.api.
		On("CheckScope", api.Scope(nil), api.Scope{"scopeItem"}).
		Return(true)

	s.api.
		On("SecretKey").
		Return("secretKey")

	s.api.
		On("CreateToken", "entityID", []string(nil), time.Second*5).
		Return(at)

	s.api.
		On("CreateToken", "entityID_refresh", []string(nil), time.Second*10).
		Return(rt)

	ctx := context.WithValue(context.Background(), "crd", credentials.Credentials{
		ID:       "entityID",
		Password: "password",
	})

	r, err := s.handler.Authenticate(ctx, connect.NewRequest(&v1.AuthenticateRequest{
		Scope: []string{"scopeItem"},
	}))
	s.Require().NoError(err)

	s.Assert().Equal("accessTokenSignedString", r.Msg.AccessToken)
	s.Assert().Equal(int64(123456789), r.Msg.AccessTokenExpires)
	s.Assert().Equal("refreshTokenSignedString", r.Msg.RefreshToken)
	s.Assert().Equal(int64(234567890), r.Msg.RefreshTokenExpires)

	l := s.logger.LastEntry()
	s.Require().NotNil(l)
	s.Assert().Equal(`{"level":"info","entity_id":"entityID","access_token_expires":123456789,"refresh_token_expires":234567890,"message":"authenticated by password"}`, l.String())
}

func TestHandler_Authenticate(t *testing.T) {
	suite.Run(t, new(AuthenticateTestSuite))
}
