package sqlparse

import (
	"bufio"
	"io"
	"strings"
)

type SqlParser struct {
	reader io.Reader
}

type ExecutionOptions struct {
	UseTransaction bool
}

func NewSqlParser(reader io.Reader) *SqlParser {
	return &SqlParser{reader: reader}
}

func (sp *SqlParser) Parse() (*ExecutionOptions, error) {
	options := &ExecutionOptions{UseTransaction: true}

	scanner := bufio.NewScanner(sp.reader)
	if scanner.Scan() {
		firstLine := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if strings.HasPrefix(firstLine, "-- bolt:") {
			parts := strings.Split(firstLine, " ")
			// Skip the first two parts ("--" and "bolt:")
			for _, part := range parts[2:] {
				if part == "no-transaction" {
					options.UseTransaction = false
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return options, nil
}
