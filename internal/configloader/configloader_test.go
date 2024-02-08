package configloader_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/checkmate/assert"
	"github.com/eugenetriguba/checkmate/check"
)

func TestNewConfigDefaults(t *testing.T) {
	bolttest.ChangeCwd(t, os.TempDir())

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)

	check.Equal(t, cfg.Migrations.DirectoryPath, "migrations")
	check.Equal(t, cfg.Migrations.VersionStyle, configloader.VersionStyleSequential)
}

func TestNewConfigWithInvalidVersionStyle(t *testing.T) {
	fileCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "myfancymigrations",
			VersionStyle:  "invalid",
		},
	}
	bolttest.CreateConfigFile(t, &fileCfg, "bolt.toml")

	_, err := configloader.NewConfig()
	assert.ErrorIs(t, err, configloader.ErrInvalidVersionStyle)
}

func TestNewConfigFindsFileAndPopulatesConfigStruct(t *testing.T) {
	expectedCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "myfancymigrations",
			VersionStyle:  configloader.VersionStyleSequential,
		},
		Connection: configloader.ConnectionConfig{
			Host:     "testhost",
			Port:     1234,
			User:     "testuser",
			Password: "testpassword",
			DBName:   "testdb",
			Driver:   "postgres",
		},
	}
	bolttest.CreateConfigFile(t, &expectedCfg, "bolt.toml")

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)
	assert.DeepEqual(t, *cfg, expectedCfg)
}

func TestNewConfigCanBeOverridenByEnvVars(t *testing.T) {
	fileCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "cfgmigrations",
			VersionStyle:  configloader.VersionStyleSequential,
		},
		Connection: configloader.ConnectionConfig{
			Host:     "testhost",
			Port:     1234,
			User:     "testuser",
			Password: "testpassword",
			DBName:   "testdb",
			Driver:   "mysql",
		},
	}
	bolttest.CreateConfigFile(t, &fileCfg, "bolt.toml")

	envCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "envmigrations",
			VersionStyle:  configloader.VersionStyleTimestamp,
		},
		Connection: configloader.ConnectionConfig{
			Host:     "envtesthost",
			Port:     4321,
			User:     "envtestuser",
			Password: "envtestpassword",
			DBName:   "envtestdb",
			Driver:   "postgres",
		},
	}
	t.Setenv("BOLT_MIGRATIONS_VERSION_STYLE", string(envCfg.Migrations.VersionStyle))
	t.Setenv("BOLT_MIGRATIONS_DIR_PATH", envCfg.Migrations.DirectoryPath)
	t.Setenv("BOLT_DB_CONN_HOST", envCfg.Connection.Host)
	t.Setenv("BOLT_DB_CONN_PORT", fmt.Sprintf("%d", envCfg.Connection.Port))
	t.Setenv("BOLT_DB_CONN_USER", envCfg.Connection.User)
	t.Setenv("BOLT_DB_CONN_PASSWORD", envCfg.Connection.Password)
	t.Setenv("BOLT_DB_CONN_DBNAME", envCfg.Connection.DBName)
	t.Setenv("BOLT_DB_CONN_DRIVER", envCfg.Connection.Driver)

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)
	assert.DeepEqual(t, *cfg, envCfg)
}
