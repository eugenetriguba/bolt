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

	check.Equal(t, cfg.Source.Filesystem.DirectoryPath, "migrations")
	check.Equal(t, cfg.Source.VersionStyle, configloader.VersionStyleTimestamp)
	check.Equal(t, cfg.Database.MigrationsTable, "bolt_migrations")
}

func TestNewConfigWithInvalidVersionStyle(t *testing.T) {
	fileCfg := configloader.Config{
		Source: configloader.SourceConfig{
			VersionStyle: "invalid",
			Filesystem: configloader.FilesystemSourceConfig{
				DirectoryPath: "myfancymigrations",
			},
		},
	}
	bolttest.CreateConfigFile(t, &fileCfg, "bolt.toml")

	_, err := configloader.NewConfig()
	assert.ErrorIs(t, err, configloader.ErrInvalidVersionStyle)
}

func TestNewConfigFindsFileAndPopulatesConfigStruct(t *testing.T) {
	bolttest.UnsetEnv(t, "BOLT_DB_DSN")
	bolttest.UnsetEnv(t, "BOLT_DB_MIGRATIONS_TABLE")
	bolttest.UnsetEnv(t, "BOLT_SOURCE_VERSION_STYLE")
	bolttest.UnsetEnv(t, "BOLT_SOURCE_FS_DIR_PATH")
	expectedCfg := configloader.Config{
		Source: configloader.SourceConfig{
			VersionStyle: configloader.VersionStyleSequential,
			Filesystem: configloader.FilesystemSourceConfig{
				DirectoryPath: "myfancymigrations",
			},
		},
		Database: configloader.DatabaseConfig{
			DSN:             "postgresql://testuser:testpassword@testhost:1234/testdb",
			MigrationsTable: "test_table",
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
		Source: configloader.SourceConfig{
			VersionStyle: configloader.VersionStyleSequential,
			Filesystem: configloader.FilesystemSourceConfig{
				DirectoryPath: "cfgmigrations",
			},
		},
		Database: configloader.DatabaseConfig{
			DSN:             "mysql://testuser:testpassword@testhost:1234/testdb",
			MigrationsTable: "test_table",
		},
	}
	bolttest.CreateConfigFile(t, &fileCfg, "bolt.toml")

	envCfg := configloader.Config{
		Source: configloader.SourceConfig{
			VersionStyle: configloader.VersionStyleTimestamp,
			Filesystem: configloader.FilesystemSourceConfig{
				DirectoryPath: "envmigrations",
			},
		},
		Database: configloader.DatabaseConfig{
			DSN:             "postgresql://envtestuser:envtestpassword@envtesthost:4321/envtestdb",
			MigrationsTable: "different_table",
		},
	}
	t.Setenv("BOLT_SOURCE_VERSION_STYLE", string(envCfg.Source.VersionStyle))
	t.Setenv("BOLT_SOURCE_FS_DIR_PATH", envCfg.Source.Filesystem.DirectoryPath)
	t.Setenv("BOLT_DB_DSN", envCfg.Database.DSN)
	t.Setenv("BOLT_DB_MIGRATIONS_TABLE", envCfg.Database.MigrationsTable)

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)
	assert.DeepEqual(t, *cfg, envCfg)
}

func TestNewConfigSearchesParentDirectories(t *testing.T) {
	bolttest.UnsetEnv(t, "BOLT_DB_DSN")
	bolttest.UnsetEnv(t, "BOLT_DB_MIGRATIONS_TABLE")
	bolttest.UnsetEnv(t, "BOLT_SOURCE_FS_DIR_PATH")
	bolttest.UnsetEnv(t, "BOLT_SOURCE_VERSION_STYLE")
	expectedCfg := configloader.Config{
		Source: configloader.SourceConfig{
			VersionStyle: configloader.VersionStyleSequential,
			Filesystem: configloader.FilesystemSourceConfig{
				DirectoryPath: "differentmigrationsdir",
			},
		},
		Database: configloader.DatabaseConfig{
			MigrationsTable: "migration_table",
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
