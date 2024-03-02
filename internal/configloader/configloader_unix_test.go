//go:build unix

package configloader_test

import (
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/checkmate/assert"
)

func TestNewConfigUnixSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{
			DirectoryPath: "differentmigrationsdir",
			VersionStyle:  configloader.VersionStyleSequential,
		},
	}
	bolttest.CreateConfigFile(t, &expectedCfg, "/bolt.toml")

	homedir, err := os.UserHomeDir()
	assert.Nil(t, err)
	bolttest.ChangeCwd(t, homedir)

	cfg, err := configloader.NewConfig()
	assert.Nil(t, err)

	assert.DeepEqual(t, *cfg, expectedCfg)
}
