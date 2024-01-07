package configloader_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigMigrationsDirDefault(t *testing.T) {
	changeCwd(t, os.TempDir())

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
	createConfigFile(t, &expectedCfg, "bolt.toml")

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
	createConfigFile(t, &fileCfg, "bolt.toml")

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

func createConfigFile(t *testing.T, cfg *configloader.Config, filePath string) {
	f := createTempFile(t, filePath)

	encoder := toml.NewEncoder(f)
	err := encoder.Encode(cfg)
	assert.NilError(t, err)

	err = f.Close()
	assert.NilError(t, err)
}

func createTempFile(t *testing.T, filePath string) *os.File {
	f, err := os.Create(filePath)
	assert.NilError(t, err)

	t.Cleanup(func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	return f
}

func changeCwd(t *testing.T, path string) {
	dir, err := os.Getwd()
	assert.NilError(t, err)

	err = os.Chdir(path)
	assert.NilError(t, err)

	t.Cleanup(func() {
		err = os.Chdir(dir)
		if err != nil {
			t.Fatal(err)
		}
	})
}
