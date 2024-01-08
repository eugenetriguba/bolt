// go:build windows

package configloader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigWindowsSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "differentmigrationsdir",
	}
	configPath := filepath.Join(os.Getenv("SystemDrive"), "bolt.toml")
	bolttest.CreateConfigFile(t, &expectedCfg, configPath)
	assert.NilError(t, err)

	homedir, err := os.UserHomeDir()
	assert.NilError(t, err)
	bolttest.ChangeCwd(t, homedir)

	cfg, err := configloader.NewConfig()
	assert.NilError(t, err)

	assert.DeepEqual(t, *cfg, expectedCfg)
}
