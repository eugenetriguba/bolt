package configloader_test

import (
	"os"
	"path/filepath"
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
	bolttest.UnsetEnv(t, "BOLT_DB_HOST")
	bolttest.UnsetEnv(t, "BOLT_DB_PORT")
	bolttest.UnsetEnv(t, "BOLT_DB_USER")
	bolttest.UnsetEnv(t, "BOLT_DB_PASSWORD")
	bolttest.UnsetEnv(t, "BOLT_DB_NAME")
	bolttest.UnsetEnv(t, "BOLT_DB_DRIVER")
	bolttest.UnsetEnv(t, "BOLT_MIGRATIONS_DIR_PATH")
	bolttest.UnsetEnv(t, "BOLT_MIGRATIONS_VERSION_STYLE")
	expectedCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "myfancymigrations",
			VersionStyle:  configloader.VersionStyleSequential,
		},
		Connection: configloader.ConnectionConfig{
			Host:     "testhost",
			Port:     "1234",
			User:     "testuser",
			Password: "testpassword",
			DBName:   "testdb",
			Driver:   "postgresql",
		},
	}
	tmpdir := t.TempDir()
	bolttest.ChangeCwd(t, tmpdir)
	bolttest.CreateConfigFile(t, &expectedCfg, filepath.Join(tmpdir, "bolt.toml"))

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
			Port:     "1234",
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
			Port:     "4321",
			User:     "envtestuser",
			Password: "envtestpassword",
			DBName:   "envtestdb",
			Driver:   "postgresql",
		},
	}
	t.Setenv("BOLT_MIGRATIONS_VERSION_STYLE", string(envCfg.Migrations.VersionStyle))
	t.Setenv("BOLT_MIGRATIONS_DIR_PATH", envCfg.Migrations.DirectoryPath)
	t.Setenv("BOLT_DB_HOST", envCfg.Connection.Host)
	t.Setenv("BOLT_DB_PORT", envCfg.Connection.Port)
	t.Setenv("BOLT_DB_USER", envCfg.Connection.User)
	t.Setenv("BOLT_DB_PASSWORD", envCfg.Connection.Password)
	t.Setenv("BOLT_DB_NAME", envCfg.Connection.DBName)
	t.Setenv("BOLT_DB_DRIVER", envCfg.Connection.Driver)

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)
	assert.DeepEqual(t, *cfg, envCfg)
}

func TestNewConfigSearchesParentDirectories(t *testing.T) {
	bolttest.UnsetEnv(t, "BOLT_DB_HOST")
	bolttest.UnsetEnv(t, "BOLT_DB_PORT")
	bolttest.UnsetEnv(t, "BOLT_DB_USER")
	bolttest.UnsetEnv(t, "BOLT_DB_PASSWORD")
	bolttest.UnsetEnv(t, "BOLT_DB_NAME")
	bolttest.UnsetEnv(t, "BOLT_DB_DRIVER")
	bolttest.UnsetEnv(t, "BOLT_MIGRATIONS_DIR_PATH")
	bolttest.UnsetEnv(t, "BOLT_MIGRATIONS_VERSION_STYLE")
	expectedCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "differentmigrationsdir",
			VersionStyle:  configloader.VersionStyleSequential,
		},
	}
	tmpdir := t.TempDir()
	bolttest.CreateConfigFile(t, &expectedCfg, filepath.Join(tmpdir, "bolt.toml"))
	nestedTmpDir := filepath.Join(tmpdir, "nested-dir", "nested-x2-dir")
	err := os.MkdirAll(nestedTmpDir, 0755)
	assert.Nil(t, err)
	bolttest.ChangeCwd(t, nestedTmpDir)

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)

	assert.DeepEqual(t, *cfg, expectedCfg)
}
