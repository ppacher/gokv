package main

import (
	"bytes"
	"fmt"

	"github.com/nethack42/gokv"
)

type listOutput struct{}

func (out listOutput) Node(n kv.Node) []byte {
	res := new(bytes.Buffer)

	fmt.Fprintln(res, "/"+n.Key)

	for _, child := range n.Children {
		res.Write(out.Node(child))
	}

	return res.Bytes()
}

func init() {
	registerOutput("list", "l", "Only display the nodes children as a list", listOutput{})
}
