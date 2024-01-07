package configloader_test

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigMigrationsDirDefault(t *testing.T) {
	cfg, err := configloader.NewConfig()
	if err != nil {
		t.Fatalf("%s occurred during new config", err)
	}

	assert.Equal(t, cfg.MigrationsDir, "migrations")
}

func TestNewConfigFindsFileAndPopulatesConfigStruct(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "migrations",
		Connection: configloader.ConnectionConfig{
			Host:     "testhost",
			Port:     1234,
			User:     "testuser",
			Password: "testpassword",
			DBName:   "testdb",
			Driver:   "postgres",
		},
	}
	err := createConfigFile(t, &expectedCfg, "bolt.toml")
	if err != nil {
		t.Fatalf("%s occurred during config file creation", err)
	}

	cfg, err := configloader.NewConfig()
	if err != nil {
		t.Fatalf("%s occurred during new config", err)
	}

	assert.DeepEqual(t, *cfg, expectedCfg)
}

func createConfigFile(t *testing.T, cfg *configloader.Config, filePath string) error {
	f, err := createTempFile(t, filePath)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func createTempFile(t *testing.T, filePath string) (*os.File, error) {
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	return f, nil
}
