package main

import "github.com/nethack42/gokv"

type valueOutput struct{}

func (out valueOutput) Node(n kv.Node) []byte {
	return n.Value
}

func init() {
	registerOutput("value", "o", "Only display the nodes value", valueOutput{})
}
