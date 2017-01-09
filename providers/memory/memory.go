package memory

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nethack42/gokv"
	"golang.org/x/net/context"
)

type Node struct {
	kv.Node

	m []*Node
}

type KV struct {
	lock sync.RWMutex

	base Node
}

func (kv *KV) Set(ctx context.Context, key string, value []byte) error {
	kv.lock.Lock()
	defer kv.lock.Unlock()

	node, err := kv.resolvePath(key, true)
	if err != nil {
		return err
	}

	node.Value = value

	return nil
}

func sanatizePath(path string) string {
	return strings.Trim(path, "/")
}

func (k *KV) resolvePath(path string, create bool) (*Node, error) {
	path = sanatizePath(path)

	parts := strings.Split(path, "/")

	k.base.IsDir = true

	node := &k.base

L:
	for i := range parts {
		key := strings.Join(parts[:i+1], "/")
		for _, child := range node.m {
			if child.Key == key {
				node = child
				continue L
			}
		}

		if !create {
			// we did not find this one
			return nil, fmt.Errorf("%q does not exist", key)
		}

		if create && !node.IsDir {
			return nil, fmt.Errorf("cannot create %q: %q is not a directory", key, node.Key)
		}

		newNode := &Node{
			Node: kv.Node{
				Key:   key,
				IsDir: i != len(parts)-1,
			},
		}

		node.m = append(node.m, newNode)
		node = newNode
		fmt.Printf("%s: created (dir=%v value=%dBytes)\n", node.Key, node.IsDir, len(node.Value))
	}

	return node, nil
}

func convertNode(n *Node) *kv.Node {
	var res kv.Node

	res = n.Node
	for _, child := range n.m {
		res.Children = append(res.Children, *convertNode(child))
	}

	return &res
}

func (k *KV) get(ctx context.Context, key string, recurse bool) (*kv.Node, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	node, err := k.resolvePath(key, false)
	if err != nil {
		return nil, err
	}

	if !recurse {
		// If we should have searched without recursion, simulate it by removing
		// references from childs
		var childs []kv.Node

		for _, c := range node.Children {
			c.Children = nil

			childs = append(childs, c)
		}

		node.Children = childs
	}

	return convertNode(node), nil
}

func (kv *KV) Get(ctx context.Context, key string) (*kv.Node, error) {
	return kv.get(ctx, key, false)
}

func (kv *KV) RGet(ctx context.Context, key string) (*kv.Node, error) {
	return kv.get(ctx, key, true)
}

func clear(node *Node) {
	for _, child := range node.m {
		clear(child)
	}

	node.m = nil
	node.Children = nil
}

func (kv *KV) Delete(ctx context.Context, key string) error {
	kv.lock.Lock()
	defer kv.lock.Unlock()

	key = sanatizePath(key)
	path := strings.Split(key, "/")
	parent := path[:len(path)-1]

	var node *Node
	var err error

	if len(parent) > 0 {
		node, err = kv.resolvePath(strings.Join(parent, "/"), false)
	} else {
		node = &kv.base
	}

	if err != nil {
		return err
	}

	for i, child := range node.m {
		if child.Key == key {
			node.m = append(node.m[:i], node.m[i+1:]...)
			clear(child)
			return nil
		}
	}

	return fmt.Errorf("%q does not exist", key)
}

func (kv *KV) CAS(ctx context.Context, key string, compare, value []byte) error {
	kv.lock.Lock()
	defer kv.lock.Unlock()

	return nil
}

func New(params map[string]string) (kv.Provider, error) {
	return &KV{}, nil
}

func init() {
	if err := kv.Register("memory", New, nil, nil); err != nil {
		panic("failed to register memory KV driver")
	}
}
