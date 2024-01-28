// go:build unix

package configloader_test

import (
	"os"
	"testing"

	"github.com/eugenetriguba/bolt/internal/bolttest"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/checkmate"
)

func TestNewConfigUnixSearchesToRootFilePath(t *testing.T) {
	expectedCfg := configloader.Config{
		Migrations: configloader.MigrationsConfig{DirectoryPath: "differentmigrationsdir"},
	}
	bolttest.CreateConfigFile(t, &expectedCfg, "/bolt.toml")

	homedir, err := os.UserHomeDir()
	checkmate.AssertNil(t, err)
	bolttest.ChangeCwd(t, homedir)

	cfg, err := configloader.NewConfig()
	checkmate.AssertNil(t, err)

	checkmate.AssertDeepEqual(t, *cfg, expectedCfg)
}
