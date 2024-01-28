package sqlparse_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/eugenetriguba/bolt/internal/sqlparse"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestSqlParser_Parse(t *testing.T) {
	testCases := []struct {
		buffer                   string
		expectedExecutionOptions *sqlparse.ExecutionOptions
	}{
		{
			buffer:                   "-- bolt: no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: false},
		},
		{
			buffer:                   "-- BOLT: no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: false},
		},
		{
			buffer:                   "\n-- bolt: no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: true},
		},
		{
			buffer:                   "",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: true},
		},
		{
			buffer:                   "--bolt: no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: true},
		},
		{
			buffer:                   "-- bolt : no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: true},
		},
		{
			buffer:                   "-- bolt: no-transaction no-transaction",
			expectedExecutionOptions: &sqlparse.ExecutionOptions{UseTransaction: false},
		},
	}

	for _, tc := range testCases {
		sqlParser := sqlparse.NewSqlParser(strings.NewReader(tc.buffer))

		execOptions, err := sqlParser.Parse()
		assert.Nil(t, err)

		assert.DeepEqual(t, execOptions, tc.expectedExecutionOptions)
	}
}

type ErrReader struct {
	Reader  io.Reader
	ErrCond string
}

func (er *ErrReader) Read(p []byte) (int, error) {
	n, err := er.Reader.Read(p)
	if err != nil {
		return n, err
	}

	if bytes.Contains(p[:n], []byte(er.ErrCond)) {
		return n, errors.New("error: unwanted input encountered")
	}

	return n, nil
}

func TestSqlParserParseError(t *testing.T) {
	reader := &ErrReader{
		Reader:  strings.NewReader("    "),
		ErrCond: " ",
	}
	sqlParser := sqlparse.NewSqlParser(reader)
	_, err := sqlParser.Parse()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unwanted input encountered")
}
