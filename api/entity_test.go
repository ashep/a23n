package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ashep/a23n/api"
	"github.com/ashep/a23n/sqldb"
)

type EntityTestSuite struct {
	suite.Suite

	db  *sqldb.DBMock
	api *api.DefaultAPI
}

func (s *EntityTestSuite) SetupTest() {
	s.db = &sqldb.DBMock{}
	s.api = api.NewDefault(s.db, "abc")
}

func (s *EntityTestSuite) TearDownTest() {
	s.db.AssertExpectations(s.T())
}

func (s *EntityTestSuite) TestCreateEntityEmptyScope() {
	_, err := s.api.CreateEntity(context.Background(), nil, nil, "aSecret", nil, nil)

	s.Require().EqualError(err, "empty scope")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestCreateEntityPasswordHashGeneratorError() {
	_, err := s.api.CreateEntity(context.Background(), nil, phgWithError(), "aSecret", api.Scope{"foo"}, nil)

	s.Require().EqualError(err, "failed to hash a secret: thePasswordHashGeneratorError")
}

func (s *EntityTestSuite) TestCreateEntityDbExecError() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}"),
	).Return(&sqldb.ResultMock{}, errors.New("theExecContextError"))

	_, err := s.api.CreateEntity(context.Background(), uuidGen(), phgNoError(), "aSecret", api.Scope{"foo"}, nil)

	s.Require().EqualError(err, "theExecContextError")
}

func (s *EntityTestSuite) TestCreateEntityOk() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)",
		[]interface{}{
			"theGeneratedUUID",
			[]byte("theGeneratedPasswordHash"),
			pq.StringArray{"theScope"},
			api.Attrs{"theAttrName": "theAttrValue"},
		},
	).Return(&sqldb.ResultMock{}, nil)

	e, err := s.api.CreateEntity(
		context.Background(),
		uuidGen(),
		phgNoError(),
		"aSecret",
		api.Scope{"theScope"},
		api.Attrs{"theAttrName": "theAttrValue"},
	)

	s.Require().NoError(err)
	s.Assert().Equal("theGeneratedUUID", e.ID)
	s.Assert().Equal("theGeneratedPasswordHash", e.Secret)
	s.Assert().Equal(api.Scope{"theScope"}, e.Scope)
	s.Assert().Equal(api.Attrs{"theAttrName": "theAttrValue"}, e.Attrs)
}
func TestDefaultAPI_Entity(t *testing.T) {
	suite.Run(t, new(EntityTestSuite))
}

func uuidGen() api.UUIDGenerator {
	return func() string {
		return "theGeneratedUUID"
	}
}

func phgWithError() api.PasswordHashGenerator {
	return func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("thePasswordHashGeneratorError")
	}
}

func phgNoError() api.PasswordHashGenerator {
	return func(password []byte, cost int) ([]byte, error) {
		return []byte("theGeneratedPasswordHash"), nil
	}
}
