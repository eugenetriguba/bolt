package sqlparse_test

import (
	"strings"
	"testing"

	"github.com/eugenetriguba/bolt/internal/sqlparse"
	"github.com/eugenetriguba/checkmate"
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
		checkmate.AssertNil(t, err)

		checkmate.AssertDeepEqual(t, execOptions, tc.expectedExecutionOptions)
	}
}
