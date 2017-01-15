// Package kv implements a generic Key-Value storage abstraction similar to
// golang's sql package.
package kv

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// Node represents an entry within the KV store.
type Node struct {
	// Key holds the absolute key for this node
	Key string `json:"key"`

	// IsDir is true if the current node represents a directory
	IsDir bool `json:"dir,omitempty"`

	// Children holds a list of child nodes. This field is only valid if IsDir
	// is set to true
	Children []Node `json:"childs,omitempty"`

	// Created stores the time the node has been created. This field is optional
	Created *time.Time `json:"created,omitempty"`

	// Updated stores the time the node has been updated last. This field is
	// optional
	Updated *time.Time `json:"updated,omitempty"`

	// Value holds the value of the node, if any. This field is only valid if
	// IsDir is set to false
	Value []byte `json:"value,omitempty"`
}

// Provider wraps databases providing basic KV operations. Users developing new
// providers must at least implement the Provider interface.
// Other interfaces (defined below) may be implemented by the provider itself.
// If they are not satisfied, a wrapper around basic KV operations is used to
// implement the same behaviour
type Provider interface {
	// Get retrieves the Node stored under path
	Get(context.Context, string) (*Node, error)

	// Set sets the value of node under path
	Set(context.Context, string, []byte) error

	// Delete deletes the node under path
	Delete(context.Context, string) error

	// CAS performs an atomic Compare-And-Swap operation
	CAS(context.Context, string, []byte, []byte) error
}

// KV wraps Key-Value databases and provides more enhanced access methods
type KV interface {
	// Provider provides basic KV operations
	Provider

	// RecursiveGetter allows to retrieve nodes recursively
	RecursiveGetter

	// KeyWatcher allows to watch a key for changes
	KeyWatcher

	// FileOps provides some file-like functionality like Copy or Move
	FileOps

	WithPrefix(string) KV
}

// RecursiveGetter allows to retrieve nodes recursively
type RecursiveGetter interface {
	RGet(context.Context, string) (*Node, error)
}

// KeyWatcher allows to watch a key for changes
type KeyWatcher interface {
	Watch(context.Context, string) (*Node, error)
}

// Mover supports moving a key or sub-tree to a different location
type Mover interface {
	Move(context.Context, string, string) error
}

// Copier supports copying a key or sub-tree to a different location
type Copier interface {
	Copy(context.Context, string, string) error
}

// FileOps provides some file-like functionality like Copy or Move
type FileOps interface {
	// Mover supports moving a key or sub-tree to a different location
	Mover

	// Copier supports copying a key or sub-tree to a different location
	Copier
}

// Factory defines a factory function for KV providers
type Factory func(map[string]string) (Provider, error)

// Open opens a new instance to a KV provider identified by name and configured
// with params
func Open(name string, params map[string]string) (KV, error) {
	lock.Lock()
	defer lock.Unlock()

	provider, ok := factories[name]

	if !ok {
		return nil, fmt.Errorf("unkown provider")
	}

	for _, key := range provider.RequiredOptions {
		if v, ok := params[key]; !ok || v == "" {
			return nil, fmt.Errorf("missing mandatory config key: %s", key)
		}
	}

	k, err := provider.F(params)
	if err != nil {
		return nil, err
	}

	return &wrapper{k}, nil
}

// ProviderEntry represents a registered KV factory function
type ProviderEntry struct {
	// F holds the Factory func
	F Factory

	// RequiredOptions holds a list of option names that must be set in the map
	// passed to F
	RequiredOptions []string

	// OptionalOptions holds a list of additional options.
	OptionalOptions []string
}

var factories map[string]ProviderEntry
var lock sync.Mutex

// Providers returns a list of registered KV providers
func Providers() map[string]ProviderEntry {
	res := make(map[string]ProviderEntry)

	lock.Lock()
	defer lock.Unlock()

	for name, p := range factories {
		res[name] = p
	}

	return res
}

// Register registers a new factory function fn using name. One can pass
// additional strings representing required configuration map keys
func Register(name string, fn Factory, required []string, optional []string) error {
	lock.Lock()
	defer lock.Unlock()

	if factories == nil {
		factories = make(map[string]ProviderEntry)
	}

	factories[name] = ProviderEntry{
		F:               fn,
		RequiredOptions: required,
		OptionalOptions: optional,
	}

	return nil
}
