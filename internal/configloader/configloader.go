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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
)

type VersionStyle string

const (
	VersionStyleSequential VersionStyle = "sequential"
	VersionStyleTimestamp  VersionStyle = "timestamp"
)

var (
	ErrConfigFileNotFound = errors.New(
		"bolt configuration file not found in current directory or any parent directories",
	)
	ErrInvalidVersionStyle = fmt.Errorf(
		"invalid version style for bolt migrations. supported styles: %v",
		[]VersionStyle{VersionStyleSequential, VersionStyleTimestamp},
	)
)

// Config represents the application configuration settings.
//
// This can come from the TOML file or environment variables,
// with environment variables taking precedence.
type Config struct {
	Migrations MigrationsConfig `toml:"migrations"`

	// Information related to how to connect to the database
	// that is desired to run migrations against.
	Connection ConnectionConfig `toml:"connection"`
}

type MigrationsConfig struct {
	DirectoryPath string       `toml:"directory_path" envconfig:"BOLT_MIGRATIONS_DIR_PATH"`
	VersionStyle  VersionStyle `toml:"version_style"  envconfig:"BOLT_MIGRATIONS_VERSION_STYLE"`
}

type ConnectionConfig struct {
	Host     string `toml:"host"     envconfig:"BOLT_DB_CONN_HOST"`
	Port     string `toml:"port"     envconfig:"BOLT_DB_CONN_PORT"`
	User     string `toml:"user"     envconfig:"BOLT_DB_CONN_USER"`
	Password string `toml:"password" envconfig:"BOLT_DB_CONN_PASSWORD"`
	DBName   string `toml:"dbname"   envconfig:"BOLT_DB_CONN_DBNAME"`
	Driver   string `toml:"driver"   envconfig:"BOLT_DB_CONN_DRIVER"`
}

func NewConfig() (*Config, error) {
	filePath, err := findConfigFilePath()
	if err != nil && !errors.Is(err, ErrConfigFileNotFound) {
		return nil, err
	}

	cfg := Config{
		Migrations: MigrationsConfig{
			DirectoryPath: "migrations",
			VersionStyle:  VersionStyleSequential,
		},
	}
	if !errors.Is(err, ErrConfigFileNotFound) {
		_, err = toml.DecodeFile(filePath, &cfg)
		if err != nil {
			return nil, err
		}
	}

	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Migrations.VersionStyle != VersionStyleSequential &&
		cfg.Migrations.VersionStyle != VersionStyleTimestamp {
		return nil, ErrInvalidVersionStyle
	}

	return &cfg, nil
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
			return "", ErrConfigFileNotFound
		}

		dir = filepath.Dir(dir)
	}
}

// fsRootDir retrieves the root directory
// of the filesystem on Windows or any unix-like
// operating system.
func fsRootDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("SystemDrive")
	}
	return "/"
}
