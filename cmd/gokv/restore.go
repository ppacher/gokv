package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/nethack42/gokv"
	"gopkg.in/urfave/cli.v2"
)

func restoreTree(c *cli.Context) error {
	path := c.Args().Get(0)

	expand := c.Bool("rel")

	var r io.Reader

	out := c.String("input")
	switch out {
	case "-":
		r = os.Stdin
	case "":
		r = os.Stdin
	default:
		f, err := os.Open(out)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	var nodes node

	if err := json.NewDecoder(r).Decode(&nodes); err != nil {
		return fmt.Errorf("failed to load json: %s", err)
	}

	prefix := strings.Trim(path, "/ ")

	tree := convertNodeFromJson(nodes)

	if expand {
		tree = expandNode(tree, prefix)
	}

	store, err := getKV(c)
	if err != nil {
		return err
	}

	saveNode(store, tree)

	return nil
}

func expandNode(n kv.Node, p string) kv.Node {
	n.Key = p + n.Key

	var childs []kv.Node

	for _, child := range n.Children {
		childs = append(childs, expandNode(child, p))
	}

	n.Children = childs

	return n
}

func saveNode(store kv.KV, n kv.Node) {
	if n.IsDir {
		for _, child := range n.Children {
			saveNode(store, child)
		}
	} else {
		store.Set(context.Background(), n.Key, n.Value)
	}
}
