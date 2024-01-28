package bolttest

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/eugenetriguba/bolt/internal/configloader"
	"github.com/eugenetriguba/checkmate/assert"
)

func CreateConfigFile(t *testing.T, cfg *configloader.Config, filePath string) {
	f := CreateTempFile(t, filePath)

	encoder := toml.NewEncoder(f)
	err := encoder.Encode(cfg)
	assert.Nil(t, err)

	err = f.Close()
	assert.Nil(t, err)
}

func CreateTempFile(t *testing.T, filePath string) *os.File {
	f, err := os.Create(filePath)
	assert.Nil(t, err)

	t.Cleanup(func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	return f
}

func ChangeCwd(t *testing.T, path string) {
	dir, err := os.Getwd()
	assert.Nil(t, err)

	err = os.Chdir(path)
	assert.Nil(t, err)

	t.Cleanup(func() {
		err = os.Chdir(dir)
		if err != nil {
			t.Fatal(err)
		}
	})
}
