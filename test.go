package kv

import (
	"testing"

	"golang.org/x/net/context"
)

func KVTester(t *testing.T, kv KV) {
	flatTests(t, kv)
	dirTests(t, kv)
}

func dirTests(t *testing.T, kv KV) {
	ctx := context.Background()

	if _, err := kv.Get(ctx, "/x/b/c"); err == nil {
		t.Errorf("kv: (dir-tests) Get() of non-existent key did not return an error")
	}

	if node, _ := kv.Get(ctx, "/x/b/c"); node != nil {
		t.Errorf("kv: (dir-tests) Get() of non-existent key did return a node")
	}

	if err := kv.Delete(ctx, "/x/b/c"); err == nil {
		t.Errorf("kv: (dir-tests) Delete() of non-existent key should return an error")
	}

	if err := kv.Set(ctx, "/a", []byte("test")); err != nil {
		t.Errorf("kv: (dir-tests) Set() returned error: %s", err)
	}

	if err := kv.Set(ctx, "/a/b", []byte("test")); err == nil {
		t.Errorf("kv: (dir-tests) Set() should fail on /a/b as /a is a file")
	}

	if err := kv.Delete(ctx, "/a"); err != nil {
		t.Errorf("kv: (dir-tests) Delete() of existent key returned error: %s", err)
	}

	if err := kv.Set(ctx, "/a/a", []byte("1")); err != nil {
		t.Errorf("kv: (dir-tests) Set() returned error: %s", err)
	}

	if err := kv.Set(ctx, "/a/b", []byte("2")); err != nil {
		t.Errorf("kv: (dir-tests) Set() returned error: %s", err)
	}

	if err := kv.Set(ctx, "/a/c/c/b", []byte("3")); err != nil {
		t.Errorf("kv: (dir-tests) Set() returned error: %s", err)
	}

	if node, err := kv.Get(ctx, "/a/a"); err != nil {
		t.Errorf("kv: (dir-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (dir-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "1" {
			t.Errorf("kv: (dir-tests) Get() returned invalid value for node")
		}
	}

	if node, err := kv.Get(ctx, "/a/b"); err != nil {
		t.Errorf("kv: (dir-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (dir-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "2" {
			t.Errorf("kv: (dir-tests) Get() returned invalid value for node")
		}
	}

	if node, err := kv.Get(ctx, "/a/c/c/b"); err != nil {
		t.Errorf("kv: (dir-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (dir-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "3" {
			t.Errorf("kv: (dir-tests) Get() returned invalid value for node")
		}
	}

	if node, err := kv.Get(ctx, "/a"); err != nil {
		t.Errorf("kv: (dir-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if !node.IsDir {
			t.Errorf("kv: (dir-tests) Get() returned Value node. Directory expected")
		}
		if node.Value != nil {
			t.Errorf("kv: (dir-tests) Get() returned invalid value for node. Should be nil but is %v", node.Value)
		}

		// we check for < since previous tests may not have cleaned up the root
		// KV space
		if len(node.Children) < 3 {
			t.Errorf("kv: (dir-tests) Get() returned invalid number of children for node: %v", node.Children)
		}

		var a, b, c bool

		for _, child := range node.Children {
			if child.Key == "a/a" {
				if a {
					t.Errorf("kv: (dir-tests) node has duplicate child")
				}
				if child.IsDir {
					t.Errorf("kv: (dir-tests) child /a/a should be value, not directory")
				}
				a = true
			}

			if child.Key == "a/b" {
				if b {
					t.Errorf("kv: (dir-tests) node has duplicate child")
				}
				if child.IsDir {
					t.Errorf("kv: (dir-tests) child /a/b should be value, not directory")
				}
				b = true
			}

			if child.Key == "a/c" {
				if c {
					t.Errorf("kv: (dir-tests) node has duplicate child")
				}
				if !child.IsDir {
					t.Errorf("kv: (dir-tests) child /a/c should be directory, not value")
				}
				c = true
			}
		}

		if !a || !b || !c {
			t.Errorf("kv: (dir-tests) Get() returned node with some childs missing (a=%v b=%v c=%v)", a, b, c)
		}
	}
}

func flatTests(t *testing.T, kv KV) {
	ctx := context.Background()

	// Get Non-Existent keys
	if _, err := kv.Get(ctx, "foobar"); err == nil {
		t.Errorf("kv: (flat-tests) Get() of non-existent key did not return an error")
	}

	if node, _ := kv.Get(ctx, "foobar"); node != nil {
		t.Errorf("kv: (flat-tests) Get() of non-existent key did return a Node")
	}

	// Add a key and retriev it with and without a leading /
	if err := kv.Set(ctx, "foobar", []byte("foobar")); err != nil {
		t.Errorf("kv: (flat-tests) Set() returned error: %s", err)
	}

	if node, err := kv.Get(ctx, "foobar"); err != nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (flat-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "foobar" {
			t.Errorf("kv: (flat-tests) Get() returned invalid value for node")
		}
	}

	if node, err := kv.Get(ctx, "/foobar"); err != nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (flat-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "foobar" {
			t.Errorf("kv: (flat-tests) Get() returned invalid value for node")
		}
	}

	// Update key using a leading /
	if err := kv.Set(ctx, "/foobar", []byte("barfoo")); err != nil {
		t.Errorf("kv: (flat-tests) Set() should succeedd (leading /)")
	}

	if node, err := kv.Get(ctx, "foobar"); err != nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned error: %s", err)
	} else if node == nil {
		t.Errorf("kv: (flat-tests) Get() of existent key returned no node")
	} else {
		if node.IsDir {
			t.Errorf("kv: (flat-tests) Get() returned directory node. Value expected")
		}
		if string(node.Value) != "barfoo" {
			t.Errorf("kv: (flat-tests) Get() returned invalid value for nodel")
		}
	}

	// Try delete key (without leading /)
	if err := kv.Delete(ctx, "foobar"); err != nil {
		t.Errorf("kv: (flat-tests) Delete() of existent key returned error: %s", err)
	}

	// add another one
	if err := kv.Set(ctx, "barfoo", []byte("barfoo")); err != nil {
		t.Errorf("kv: (flat-tests) Set() returned error: %s", err)
	}

	if err := kv.Delete(ctx, "/barfoo"); err != nil {
		t.Errorf("kv: (flat-tests) Delete() of existent key should succeeed (leading /)")
	}

	if err := kv.Delete(ctx, "does-not-exist"); err == nil {
		t.Errorf("kv: (flat-tests) Delete() of non-existent key should fail")
	}
}
