package main

import (
	"bytes"
	"encoding/json"

	"github.com/nethack42/gokv"
)

type jsonOutput struct{}

type node struct {
	kv.Node

	Value    string `json:"value,omitempty"`
	Children []node `json:"childs,omitempty"`
}

func convertNodeFromJson(n node) kv.Node {
	res := n.Node
	res.Value = []byte(n.Value)

	for _, child := range n.Children {
		res.Children = append(res.Children, convertNodeFromJson(child))
	}

	return res
}

func convertNodeToJson(n kv.Node) node {
	res := node{
		Node:  n,
		Value: string(n.Value),
	}

	for _, child := range n.Children {
		res.Children = append(res.Children, convertNodeToJson(child))
	}

	return res
}

func (out jsonOutput) Node(n kv.Node) []byte {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(convertNodeToJson(n))

	return buf.Bytes()
}

func init() {
	registerOutput("json", "j", "Display result as JSON", jsonOutput{})

	// JSON should be used as a default
	defaultOutput = jsonOutput{}
}
