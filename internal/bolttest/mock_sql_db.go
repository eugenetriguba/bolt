package bolttest

import "database/sql"

type MockSqlDb struct {
	ExecReturnValue     ExecReturnValue
	ExecCallCount       int
	QueryReturnValue    QueryReturnValue
	QueryCallCount      int
	QueryRowReturnValue QueryRowReturnValue
	QueryRowCallCount   int
	BeginReturnValue    BeginReturnValue
	BeginCallCount      int
}

type ExecReturnValue struct {
	Result sql.Result
	Err    error
}

type QueryReturnValue struct {
	Rows *sql.Rows
	Err  error
}

type QueryRowReturnValue struct {
	Row *sql.Row
}

type BeginReturnValue struct {
	Tx  *sql.Tx
	Err error
}

func (db *MockSqlDb) Exec(query string, args ...any) (sql.Result, error) {
	db.ExecCallCount += 1
	return db.ExecReturnValue.Result, db.ExecReturnValue.Err
}

func (db *MockSqlDb) Query(query string, args ...any) (*sql.Rows, error) {
	db.QueryCallCount += 1
	return db.QueryReturnValue.Rows, db.QueryReturnValue.Err
}

func (db *MockSqlDb) QueryRow(query string, args ...any) *sql.Row {
	db.QueryRowCallCount += 1
	return db.QueryRowReturnValue.Row
}

func (db *MockSqlDb) Begin() (*sql.Tx, error) {
	db.BeginCallCount += 1
	return db.BeginReturnValue.Tx, db.BeginReturnValue.Err
}
