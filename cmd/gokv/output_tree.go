package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nethack42/gokv"
)

type treeOutput struct{}

func treePrintNode(w io.Writer, strip string, n kv.Node, prefix string, hasNext bool) {
	key := strings.Replace(n.Key, strip, "", -1)

	if len(key) > 0 {
		if key[0] == '/' {
			key = key[1:]
		}

		if hasNext {
			fmt.Printf("%s├── %s\n", prefix, key)
		} else {
			fmt.Printf("%s└── %s\n", prefix, key)
		}
	}

	if n.IsDir {
		for i, child := range n.Children {
			p := prefix + "   "

			if hasNext {
				p = prefix + "│  "
			}

			treePrintNode(w, n.Key, child, p, i+1 != len(n.Children))
		}
	}
}

func (out treeOutput) Node(n kv.Node) []byte {
	res := new(bytes.Buffer)

	treePrintNode(res, "", n, "", false)

	return res.Bytes()
}

func init() {
	registerOutput("tree", "t", "Only display the nodes children as a tree", treeOutput{})
}
