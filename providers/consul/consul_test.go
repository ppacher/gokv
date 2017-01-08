package consul

import (
	"testing"

	"github.com/nethack42/gokv"
)

func Test_Consul(t *testing.T) {
	k, err := New(map[string]string{})

	if err != nil {
		t.Errorf("failed to create consul KV")
		t.FailNow()
	}

	kv.KVTester(t, k)
}
