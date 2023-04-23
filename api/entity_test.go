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

func (s *EntityTestSuite) TestCreateEntityEmptyID() {
	err := s.api.CreateEntity(context.Background(), "", nil, nil, nil)

	s.Require().EqualError(err, "invalid id: invalid UUID length: 0")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestCreateEntityInvalidID() {
	err := s.api.CreateEntity(context.Background(), "de2a6f34-5371-4409-89ec-62bfda13fcby", nil, nil, nil)

	s.Require().EqualError(err, "invalid id: invalid UUID format")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestCreateEntityEmptySecret() {
	err := s.api.CreateEntity(context.Background(), "de2a6f34-5371-4409-89ec-62bfda13fcb7", nil, nil, nil)

	s.Require().EqualError(err, "invalid secret: crypto/bcrypt: hashedSecret too short to be a bcrypted password")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestCreateEntityInvalidSecret() {
	err := s.api.CreateEntity(context.Background(), "de2a6f34-5371-4409-89ec-62bfda13fcb7", []byte("abc"), nil, nil)

	s.Require().EqualError(err, "invalid secret: crypto/bcrypt: hashedSecret too short to be a bcrypted password")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestCreateEntityDbExecError() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}"),
	).Return(&sqldb.ResultMock{}, errors.New("theExecContextError"))

	err := s.api.CreateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		nil,
		nil,
	)

	s.Require().EqualError(err, "theExecContextError")
}

func (s *EntityTestSuite) TestCreateEntityEmptyScope() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)",
		[]interface{}{
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{},
			[]byte(`{"attrName":"attrValue"}`),
		},
	).Return(&sqldb.ResultMock{}, nil)

	err := s.api.CreateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		nil,
		api.Attrs{"attrName": "attrValue"},
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestCreateEntityEmptyAttrs() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)",
		[]interface{}{
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{"aScope"},
			[]byte("{}"),
		},
	).Return(&sqldb.ResultMock{}, nil)

	err := s.api.CreateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		api.Scope{"aScope"},
		nil,
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestCreateEntityEmptyScopeAndAttrs() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)",
		[]interface{}{
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{},
			[]byte("{}"),
		},
	).Return(&sqldb.ResultMock{}, nil)

	err := s.api.CreateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		nil,
		nil,
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestCreateEntityOk() {
	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"INSERT INTO entity (id, secret, scope, attrs) VALUES ($1, $2, $3, $4)",
		[]interface{}{
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{"theScope"},
			[]byte(`{"theAttrName":"theAttrValue"}`),
		},
	).Return(&sqldb.ResultMock{}, nil)

	err := s.api.CreateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		api.Scope{"theScope"},
		api.Attrs{"theAttrName": "theAttrValue"},
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestUpdateEntityEmptyID() {
	err := s.api.UpdateEntity(context.Background(), "", nil, nil, nil)

	s.Require().EqualError(err, "invalid id: invalid UUID length: 0")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestUpdateEntityInvalidID() {
	err := s.api.UpdateEntity(context.Background(), "de2a6f34-5371-4409-89ec-62bfda13fcby", nil, nil, nil)

	s.Require().EqualError(err, "invalid id: invalid UUID format")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestUpdateEntityInvalidSecret() {
	err := s.api.UpdateEntity(context.Background(), "de2a6f34-5371-4409-89ec-62bfda13fcb7", []byte("abc"), nil, nil)

	s.Require().EqualError(err, "invalid secret: crypto/bcrypt: hashedSecret too short to be a bcrypted password")
	s.Assert().ErrorIs(err, api.ErrInvalidArg{})
}

func (s *EntityTestSuite) TestUpdateEntityRowsAffectedError() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(0), errors.New("theRowsAffectedError"))

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}"),
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		nil,
		nil,
		nil,
	)

	s.Require().EqualError(err, "theRowsAffectedError")
}

func (s *EntityTestSuite) TestUpdateEntityNotFoundEntity() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(0), nil)

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}"),
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		nil,
		nil,
		nil,
	)

	s.Require().ErrorIs(err, api.ErrNotFound)
}

func (s *EntityTestSuite) TestUpdateEntityEmptySecret() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(1), nil)

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"UPDATE entity SET scope=$1, attrs=$2 WHERE id=$3",
		[]interface{}{
			pq.StringArray{},
			[]byte("{}"),
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		},
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		nil,
		nil,
		nil,
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestUpdateEntityNonEmptySecret() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(1), nil)

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"UPDATE entity SET secret=$1, scope=$2, attrs=$3 WHERE id=$4",
		[]interface{}{
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{},
			[]byte("{}"),
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		},
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		nil,
		nil,
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestUpdateEntityNonEmptyScope() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(1), nil)

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"UPDATE entity SET secret=$1, scope=$2, attrs=$3 WHERE id=$4",
		[]interface{}{
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{"theScope"},
			[]byte("{}"),
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		},
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		api.Scope{"theScope"},
		nil,
	)

	s.Require().NoError(err)
}

func (s *EntityTestSuite) TestUpdateEntityNonEmptyAttrs() {
	res := &sqldb.ResultMock{}

	res.On("RowsAffected").Return(int64(1), nil)

	s.db.On(
		"ExecContext",
		mock.AnythingOfType("*context.emptyCtx"),
		"UPDATE entity SET secret=$1, scope=$2, attrs=$3 WHERE id=$4",
		[]interface{}{
			[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
			pq.StringArray{"theScope"},
			[]byte(`{"theAttrName":"theAttrValue"}`),
			"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		},
	).Return(res, nil)

	err := s.api.UpdateEntity(
		context.Background(),
		"de2a6f34-5371-4409-89ec-62bfda13fcb7",
		[]byte("$2a$12$5GiSCPaURd2vLHGm.HgtF.SJGGPjJyuXaiTFnKSCqVbRTJl75ZUvy"),
		api.Scope{"theScope"},
		api.Attrs{"theAttrName": "theAttrValue"},
	)

	s.Require().NoError(err)
}

func TestDefaultAPI_Entity(t *testing.T) {
	suite.Run(t, new(EntityTestSuite))
}
