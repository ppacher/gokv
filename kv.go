// Package kv implements a generic Key-Value storage abstraction similar to
// golang's sql package.
package kv

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
)

// Node represents an entry within the KV store.
type Node struct {
	// Key holds the absolute key for this node
	Key string

	// IsDir is true if the current node represents a directory
	IsDir bool

	// Children holds a list of child nodes. This field is only valid if IsDir
	// is set to true
	Children []Node

	// Created stores the time the node has been created. This field is optional
	Created *time.Time

	// Updated stores the time the node has been updated last. This field is
	// optional
	Updated *time.Time

	// Value holds the value of the node, if any. This field is only valid if
	// IsDir is set to false
	Value []byte
}

// KV wraps Key-Value store providers
type KV interface {
	// Get retrieves the Node stored under path
	Get(context.Context, string) (*Node, error)

	// Set sets the value of node under path
	Set(context.Context, string, []byte) error

	// Delete deletes the node under path
	Delete(context.Context, string) error

	// CAS performs an atomic Compare-And-Swap operation
	CAS(context.Context, string, []byte, []byte) error
}

// Factory defines a factory function for KV providers
type Factory func(map[string]string) (KV, error)

// Open opens a new instance to a KV provider identified by name and configured
// with params
func Open(name string, params interface{}) (KV, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// Register registers a new factory function fn using name. One can pass
// additional strings representing required configuration map keys
func Register(name string, fn Factory, opts ...string) error {
	return nil
}
