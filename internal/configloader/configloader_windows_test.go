//go:build windows

package configloader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigWindowsSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "differentmigrationsdir",
	}
	configPath := filepath.Join(os.Getenv("SystemDrive"), "bolt.toml")
	err := createConfigFile(t, &expectedCfg, configPath)
	if err != nil {
		t.Fatal(err)
	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(homedir)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := configloader.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	assert.DeepEqual(t, *cfg, expectedCfg)
}
