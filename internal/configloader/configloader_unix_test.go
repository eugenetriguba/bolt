//go:build unix

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
	err := createConfigFile(t, &expectedCfg, "/bolt.toml")
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
