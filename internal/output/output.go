package output

type Outputter interface {
	Output(message string) error
	Error(message string) error
	Table(header []string, rows [][]string) error
}
