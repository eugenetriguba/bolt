package output

import (
	"bytes"
	"io"
	"testing"

	"github.com/eugenetriguba/checkmate"
)

func NewConsoleOutputterWithWriters(stdout io.Writer, stderr io.Writer) ConsoleOutputter {
	return ConsoleOutputter{stdout: stdout, stderr: stderr}
}

func TestConsoleOutputter_Output(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	consoleOutputter := NewConsoleOutputterWithWriters(&stdout, &stderr)

	consoleOutputter.Output("test string")

	checkmate.AssertEqual(t, stdout.String(), "test string\n")
	checkmate.AssertEqual(t, stderr.String(), "")
}

func TestConsoleOutputter_Error(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	consoleOutputter := NewConsoleOutputterWithWriters(&stdout, &stderr)

	consoleOutputter.Error("test string")

	checkmate.AssertEqual(t, stdout.String(), "")
	checkmate.AssertEqual(t, stderr.String(), "test string\n")
}

func TestConsoleOutputter_Table(t *testing.T) {
	type test struct {
		headers        []string
		rows           [][]string
		expectedStdout string
	}

	tests := []test{
		{
			headers:        []string{},
			rows:           make([][]string, 0),
			expectedStdout: "",
		},
		{
			headers: []string{"Version"},
			rows:    [][]string{{"test1"}, {"test2"}},
			expectedStdout: "Version    \n" +
				"test1      \n" +
				"test2      \n",
		},
		{
			headers: []string{"Version", "Message"},
			rows:    [][]string{{"v1", "m1"}, {"v1", "m2"}},
			expectedStdout: "Version    Message    \n" +
				"v1         m1         \n" +
				"v1         m2         \n",
		},
	}

	for _, tc := range tests {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		consoleOutputter := NewConsoleOutputterWithWriters(&stdout, &stderr)

		consoleOutputter.Table(tc.headers, tc.rows)

		checkmate.AssertEqual(t, stdout.String(), tc.expectedStdout)
		checkmate.AssertEqual(t, stderr.String(), "")
	}
}
