// go:build unix

package configloader_test

import (
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/configloader"
	"gotest.tools/v3/assert"
)

func TestNewConfigUnixSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		MigrationsDir: "differentmigrationsdir",
	}
	createConfigFile(t, &expectedCfg, "/bolt.toml")

	homedir, err := os.UserHomeDir()
	assert.NilError(t, err)
	changeCwd(t, homedir)

	cfg, err := configloader.NewConfig()
	assert.NilError(t, err)

	assert.DeepEqual(t, *cfg, expectedCfg)
}
