package dbtest

import (
	"os"
	"strconv"

	"github.com/eugenetriguba/bolt/internal/configloader"
)

func NewTestConnectionConfig(driver string) (*configloader.ConnectionConfig, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	return &configloader.ConnectionConfig{
		Driver:   driver,
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}, nil
}
