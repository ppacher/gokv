package memory

import (
	"testing"

	"github.com/nethack42/gokv"
)

func Test_Memory(t *testing.T) {
	k, err := New(nil)

	if err != nil {
		t.Errorf("Failed to construct KV store")
		t.FailNow()
	}

	kv.KVTester(t, k)
}
