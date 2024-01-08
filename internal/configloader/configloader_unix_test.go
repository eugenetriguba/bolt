// go:build unix

package configloader_test

import (
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigUnixSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "differentmigrationsdir",
	}
	bolttest.CreateConfigFile(t, &expectedCfg, "/bolt.toml")

	homedir, err := os.UserHomeDir()
	assert.NilError(t, err)
	bolttest.ChangeCwd(t, homedir)

	cfg, err := configloader.NewConfig()
	assert.NilError(t, err)

	assert.DeepEqual(t, *cfg, expectedCfg)
}
