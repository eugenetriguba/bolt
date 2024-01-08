package configloader_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigMigrationsDirDefault(t *testing.T) {
	bolttest.ChangeCwd(t, os.TempDir())

	cfg, err := configloader.NewConfig()
	assert.NilError(t, err)

	assert.Equal(t, cfg.MigrationsDir, "migrations")
}

func TestNewConfigFindsFileAndPopulatesConfigStruct(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "myfancymigrations",
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
	assert.NilError(t, err)
	assert.DeepEqual(t, *cfg, expectedCfg)
}

func TestNewConfigCanBeOverridenByEnvVars(t *testing.T) {
	fileCfg := configloader.Config{
		MigrationsDir: "cfgmigrations",
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
		MigrationsDir: "envmigrations",
		Connection: configloader.ConnectionConfig{
			Host:     "envtesthost",
			Port:     4321,
			User:     "envtestuser",
			Password: "envtestpassword",
			DBName:   "envtestdb",
			Driver:   "postgres",
		},
	}
	t.Setenv("BOLT_MIGRATIONS_DIR", envCfg.MigrationsDir)
	t.Setenv("BOLT_CONNECTION_HOST", envCfg.Connection.Host)
	t.Setenv("BOLT_CONNECTION_PORT", fmt.Sprintf("%d", envCfg.Connection.Port))
	t.Setenv("BOLT_CONNECTION_USER", envCfg.Connection.User)
	t.Setenv("BOLT_CONNECTION_PASSWORD", envCfg.Connection.Password)
	t.Setenv("BOLT_CONNECTION_DBNAME", envCfg.Connection.DBName)
	t.Setenv("BOLT_CONNECTION_DRIVER", envCfg.Connection.Driver)

	cfg, err := configloader.NewConfig()
	assert.NilError(t, err)
	assert.DeepEqual(t, *cfg, envCfg)
}
