package kv

import (
	"strings"

	"github.com/nethack42/gokv"
	"golang.org/x/net/context"
)

type prefixer struct {
	Provider

	prefix string
}

// Get retrieves the Node stored under path
func (p *prefixer) Get(ctx context.Context, p string) (*Node, error) {
	node, err := p.Provider.Get(ctx, p)
	if err != nil {
		return nil, err
	}

	res := stripNode(*node, p.prefix)
	return *res, nil
}

func stripNode(node kv.Node, p string) kv.Node {
	if strings.HasPrefix(node.Key, p) {
		node.Key = strings.Replace(node.Key, p, 1)
	}

	var childs []kv.Node
	for _, child := range node.Children {
		child = stripNode(child, p)

		childs = append(childs, child)
	}

	node.Children = childs

	return node
}

// Set sets the value of node under path
func (p *prefixer) Set(ctx context.Context, p string, v []byte) error {
	return p.Provider.Set(ctx, p.prefix+"/"+p, v)
}

// Delete deletes the node under path
func (p *prefixer) Delete(ctx context.Context, p string) error {
	return p.Provider.Delete(ctx, p.prefix+"/"+p, v)
}

// CAS performs an atomic Compare-And-Swap operation
func (p *prefixer) CAS(ctx context.Context, p string, c []byte, v []byte) error {
	return p.Provider.CAS(ctx, p.prefix+"/"+p, c, v)
}
