// Package configloader implements a Config type for loading in the
// configuration to be used by the app.
//
// It supports reading in configuration from the "bolt.toml" file
// if it can be found in the current directory or any parent directory
// and it supports reading in environment variables. Furthermore, if
// both a configuration file is found and environment variables are set,
// the environment variables will take precedence.
package configloader

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
)

var errConfigFileNotFound = errors.New(
	"bolt configuration file not found in current directory or any parent directories",
)

// Config represents the application configuration settings.
//
// This can come from the TOML file or environment variables,
// with environment variables taking precedence.
type Config struct {
	// Relative file path from the bolt configuration file.
	// This is optional and will default to "migrations" if
	// not specified.
	MigrationsDir string `toml:"migrations_dir" envconfig:"BOLT_MIGRATIONS_DIR"`

	// Information related to how to connect to the database
	// that is desired to run migrations against.
	Connection ConnectionConfig `toml:"connection"`
}

type ConnectionConfig struct {
	Host     string `toml:"host" envconfig:"BOLT_CONNECTION_HOST"`
	Port     int    `toml:"port" envconfig:"BOLT_CONNECTION_PORT"`
	User     string `toml:"user" envconfig:"BOLT_CONNECTION_USER"`
	Password string `toml:"password" envconfig:"BOLT_CONNECTION_PASSWORD"`
	DBName   string `toml:"dbname" envconfig:"BOLT_CONNECTION_DBNAME"`
	Driver   string `toml:"driver" envconfig:"BOLT_CONNECTION_DRIVER"`
}

func NewConfig() (*Config, error) {
	filePath, err := findConfigFilePath()
	if err != nil && !errors.Is(err, errConfigFileNotFound) {
		return nil, err
	}

	cfg := &Config{}
	_, err = toml.DecodeFile(filePath, &cfg)
	if err != nil {
		return nil, err
	}

	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.MigrationsDir == "" {
		cfg.MigrationsDir = "migrations"
	}

	return cfg, nil
}

func findConfigFilePath() (filePath string, err error) {
	const configFileName = "bolt.toml"
	var rootDir = fsRootDir()

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

		if dir == rootDir {
			return "", errConfigFileNotFound
		}

		dir = filepath.Dir(dir)
	}
}

func fsRootDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("SystemDrive")
	}
	return "/"
}
