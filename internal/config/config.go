package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const configFileName = "bolt.toml"

var config *Config

type Config struct {
	// Relative file path from the bolt configuration file.
	// This is optional and will default to "migrations" if
	// not specified.
	MigrationsDir string `toml:"migrations_dir"`

	// Information related to how to connect to the database
	// that is desired to run migrations against.
	Connection connectionConfig
}

type connectionConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Driver   string
}

func NewConfig() (*Config, error) {
	if config == nil {
		filePath, err := findConfigFilePath()
		if err != nil {
			return nil, err
		}

		config = &Config{}
		_, err = toml.DecodeFile(filePath, &config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func findConfigFilePath() (filePath string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}

		for _, e := range entries {
			if e.Name() == configFileName {
				return filepath.Join(dir, e.Name()), nil
			}
		}

		// TODO: This won't support Windows.
		if dir == "/" {
			errMsg := fmt.Sprintf(
				"%s not found in current directory or any parent directories.",
				configFileName,
			)
			return "", errors.New(errMsg)
		}

		dir = filepath.Dir(dir)
	}
}
