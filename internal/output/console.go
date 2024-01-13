package output

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type ConsoleOutputter struct {
	stdout io.Writer
	stderr io.Writer
}

func NewConsoleOutputter() ConsoleOutputter {
	return ConsoleOutputter{stdout: os.Stdout, stderr: os.Stderr}
}

func (c ConsoleOutputter) Output(message string) error {
	_, err := fmt.Fprintln(c.stdout, message)
	return err
}

func (c ConsoleOutputter) Error(message string) error {
	_, err := fmt.Fprintln(c.stderr, message)
	return err
}

func (c ConsoleOutputter) Table(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(c.stdout, 4, 4, 4, ' ', 0)

	for _, header := range headers {
		_, err := fmt.Fprintf(w, "%s\t", header)
		if err != nil {
			return err
		}
	}
	if len(headers) > 0 {
		_, err := fmt.Fprint(w, "\n")
		if err != nil {
			return err
		}
	}

	for _, row := range rows {
		for _, cell := range row {
			_, err := fmt.Fprintf(w, "%s\t", cell)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "\n")
		if err != nil {
			return err
		}
	}

	w.Flush()
	return nil
}
