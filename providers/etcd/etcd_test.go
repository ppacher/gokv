package etcd

import (
	"testing"

	"github.com/nethack42/gokv"
)

func Test_Etcd(t *testing.T) {
	e, err := New(map[string]string{
		"endpoints": "http://localhost:4001/",
	})

	if err != nil {
		t.Errorf("failed to create etcd: %s", err)
		t.FailNow()
	}

	kv.KVTester(t, e)
}
