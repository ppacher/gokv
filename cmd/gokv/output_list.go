package main

import (
	"bytes"
	"fmt"

	"github.com/nethack42/gokv"
)

type listOutput struct{}

func (out listOutput) Node(n kv.Node) []byte {
	res := new(bytes.Buffer)

	for _, child := range n.Children {
		fmt.Fprintln(res, "/"+child.Key)
	}

	return res.Bytes()
}

func init() {
	registerOutput("list", "l", "Only display the nodes children as a list", listOutput{})
}
