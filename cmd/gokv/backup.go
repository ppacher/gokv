package main

import (
	"io"
	"os"

	"github.com/nethack42/gokv"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v2"
)

func backupTree(c *cli.Context) error {
	path := c.Args().Get(0)

	strip := c.Bool("rel")

	var w io.Writer

	out := c.String("output")
	switch out {
	case "stderr":
		w = os.Stderr
	case "":
		w = os.Stdout
	default:
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	store, err := getKV(c)
	if err != nil {
		return err
	}

	ctx := context.Background()

	rawTree, err := store.RGet(ctx, path)
	if err != nil {
		return err
	}

	tree := *rawTree
	if strip {
		// Strip away the the Key of node (=sanatize(path)) from all childs
		tree = stripNode(*rawTree, rawTree.Key)
	}

	w.Write(jsonOutput{}.Node(tree))
	return nil
}

func stripNode(n kv.Node, p string) kv.Node {
	n.Key = stripKey(n.Key, p)

	var childs []kv.Node

	for _, child := range n.Children {
		childs = append(childs, stripNode(child, p))
	}

	n.Children = childs

	return n
}

func stripKey(key string, prefix string) string {
	return key[len(prefix):]
}
