package sqldb

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"
)

type RowMock struct {
	mock.Mock
}

func (r *RowMock) Scan(args ...interface{}) error {
	mArgs := r.Called(args...)
	return mArgs.Error(0)
}

func (r *RowMock) Err() error {
	mArgs := r.Called()
	return mArgs.Error(0)
}

type ResultMock struct {
	mock.Mock
}

func (m *ResultMock) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *ResultMock) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

type DBMock struct {
	mock.Mock
}

func (m *DBMock) DB() *sql.DB {
	return m.Called().Get(0).(*sql.DB)
}

func (m *DBMock) PingContext(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *DBMock) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	mArgs := m.Called(ctx, query, args)
	return mArgs.Get(0).(sql.Result), mArgs.Error(1)
}

func (m *DBMock) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	mArgs := m.Called(ctx, query, args)
	return mArgs.Get(0).(Row)
}
