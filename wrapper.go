package kv

import (
	"fmt"

	"golang.org/x/net/context"
)

type wrapper struct {
	Provider
}

func (w *wrapper) fillNode(ctx context.Context, n Node) (*Node, error) {
	cur, err := w.Get(ctx, n.Key)
	if err != nil {
		return nil, err
	}

	if cur.IsDir {
		var childs []Node
		for _, child := range cur.Children {
			if child, err := w.fillNode(ctx, child); err != nil {
				return nil, err
			} else {
				childs = append(childs, *child)
			}
		}

		cur.Children = childs
	}

	return cur, nil
}

func (w *wrapper) RGet(ctx context.Context, key string) (*Node, error) {
	if v, ok := w.Provider.(RecursiveGetter); ok {
		return v.RGet(ctx, key)
	}

	root, err := w.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// fall back to get
	return w.fillNode(ctx, *root)
}

func (w *wrapper) Watch(ctx context.Context, key string) (*Node, error) {
	if v, ok := w.Provider.(KeyWatcher); ok {
		return v.Watch(ctx, key)
	}

	return nil, fmt.Errorf("Watch not supported by provider")
}

func (w *wrapper) Move(ctx context.Context, keyOld, keyNew string) error {
	if v, ok := w.Provider.(Mover); ok {
		return v.Move(ctx, keyOld, keyNew)
	}

	return fmt.Errorf("Move not supported by provider")
}

func (w *wrapper) Copy(ctx context.Context, keyOld, keyNew string) error {
	if v, ok := w.Provider.(Copier); ok {
		return v.Copy(ctx, keyOld, keyNew)
	}

	return fmt.Errorf("Copy not supported by provider")
}
