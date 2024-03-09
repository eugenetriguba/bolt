package bolttest

import (
	"os"
	"testing"

	"github.com/eugenetriguba/checkmate/assert"
)

func UnsetEnv(t *testing.T, key string) {
	originalVal, isSet := os.LookupEnv(key)
	if isSet {
		t.Cleanup(func() {
			err := os.Setenv(key, originalVal)
			assert.Nil(t, err)
		})
		os.Unsetenv(key)
	}
}
