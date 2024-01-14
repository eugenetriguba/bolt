package sqlparse_test

import (
	"strings"
	"testing"

	"github.com/eugenetriguba/bolt/internal/sqlparse"
	"gotest.tools/v3/assert"
)

func TestSqlParser_Parse(t *testing.T) {
	type test struct {
		buffer                   string
		expectedExecutionOptions *sqlparse.ExecutionOptions
	}

	tests := []test{
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

	for _, tc := range tests {
		sqlParser := sqlparse.NewSqlParser(strings.NewReader(tc.buffer))

		execOptions, err := sqlParser.Parse()
		assert.NilError(t, err)

		assert.DeepEqual(t, execOptions, tc.expectedExecutionOptions)
	}
}
