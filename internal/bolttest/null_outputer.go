package bolttest

type NullOutputter struct {
	OutputLogs []string
	ErrorLogs  []string
	TableLogs  []string
}

func (o NullOutputter) Output(message string) error {
	return nil
}

func (o NullOutputter) Error(err error) error {
	return nil
}

func (o NullOutputter) Table(header []string, rows [][]string) error {
	return nil
}
