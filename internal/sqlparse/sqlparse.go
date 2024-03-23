package sqlparse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type SqlParser interface {
	Parse(reader io.Reader) (MigrationScript, MigrationScript, error)
}

type MigrationScript struct {
	Contents string
	Options  ExecutionOptions
}

type ExecutionOptions struct {
	UseTransaction bool
}

const (
	unknownSection = iota
	upgradeScriptSection
	downgradeScriptSection
)

const (
	upgradeScriptDeliminator   = "-- migrate:up"
	downgradeScriptDeliminator = "-- migrate:down"
	transactionOptionName      = "transaction"
)

type sqlParser struct{}

// NewSqlParser creates a new SqlParser which can
// parse out a migration script.
func NewSqlParser() SqlParser {
	return &sqlParser{}
}

// Parse parses out the migration's upgrade and downgrade scripts
// and any custom execution options with those scripts. It returns
// the upgrade script, downgrade script, and any error if one occurred.
func (sp *sqlParser) Parse(reader io.Reader) (MigrationScript, MigrationScript, error) {
	upgradeScript := MigrationScript{}
	downgradeScript := MigrationScript{}

	scanner := bufio.NewScanner(reader)
	currentSection := unknownSection
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		loweredLine := strings.ToLower(line)

		if strings.HasPrefix(loweredLine, upgradeScriptDeliminator) {
			currentSection = upgradeScriptSection
			upgradeScript.Options = parseExecutionOptions(line)
		} else if strings.HasPrefix(loweredLine, downgradeScriptDeliminator) {
			currentSection = downgradeScriptSection
			downgradeScript.Options = parseExecutionOptions(line)
		} else if currentSection == upgradeScriptSection {
			upgradeScript.Contents += line + "\n"
		} else if currentSection == downgradeScriptSection {
			downgradeScript.Contents += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		return MigrationScript{}, MigrationScript{}, fmt.Errorf(
			"parsing sql file encountered an error: %w",
			err,
		)
	}

	return upgradeScript, downgradeScript, nil
}

// parseExecutionOptions extracts the execution options for a
// upgrade or downgrade migration script.
func parseExecutionOptions(line string) ExecutionOptions {
	options := ExecutionOptions{UseTransaction: true}

	parts := strings.Split(line, " ")
	for _, part := range parts {
		part = strings.ToLower(part)
		if strings.Contains(part, "transaction:") {
			option := strings.Split(part, ":")[1]
			options.UseTransaction = option != "false"
		}
	}

	return options
}
