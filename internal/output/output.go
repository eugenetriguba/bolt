package output

type Outputter interface {
	Output(message string) error
	Error(err error) error
	Table(header []string, rows [][]string) error
}
