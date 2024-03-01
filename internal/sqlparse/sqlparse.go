package sqlparse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type SqlParser interface {
	Parse(reader io.Reader) (ExecutionOptions, error)
}

type ExecutionOptions struct {
	UseTransaction bool
}

type sqlParser struct{}

func NewSqlParser() SqlParser {
	return sqlParser{}
}

func (sp sqlParser) Parse(reader io.Reader) (ExecutionOptions, error) {
	options := ExecutionOptions{UseTransaction: true}

	scanner := bufio.NewScanner(reader)
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
		return options, fmt.Errorf("parsing sql file encountered an error: %w", err)
	}

	return options, nil
}
