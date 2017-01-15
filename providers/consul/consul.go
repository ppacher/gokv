package consul

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"github.com/hashicorp/consul/api"
	"github.com/nethack42/gokv"
)

type KV struct {
	kv  *api.KV
	cli *api.Client
}

func (consul *KV) Set(ctx context.Context, key string, value []byte) error {
	v := &api.KVPair{
		Key:   key,
		Value: value,
	}

	if _, err := consul.kv.Put(v, nil); err != nil {
		return err
	}

	return nil
}

func (consul *KV) Get(ctx context.Context, key string) (*kv.Node, error) {
	pairs, _, err := consul.kv.List(key, nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%q: %#v %#v \n", key, pairs, err)

	if pairs == nil {
		return nil, errors.New("no such key")
	}

	var node kv.Node

	node.Key = key

	for _, pair := range pairs {
		var child kv.Node

		child.Key = pair.Key
		child.Value = pair.Value

		node.Children = append(node.Children, child)
	}

	return nil, nil
}

func (consul *KV) Delete(ctx context.Context, key string) error {
	if _, err := consul.kv.Delete(key, nil); err != nil {
		return err
	}
	return nil
}

func (consul *KV) CAS(ctx context.Context, key string, compare, value []byte) error {
	return nil
}

func sanatizeKey(key string) string {
	return strings.Trim(key, "/ ")
}

func New(params map[string]string) (kv.Provider, error) {
	config := api.DefaultConfig()

	if v, ok := params["endpoint"]; ok && v != "" {
		url, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		config.Address = url.Host
		config.Scheme = url.Scheme
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	kvapi := client.KV()

	return &KV{
		kv:  kvapi,
		cli: client,
	}, nil
}

func init() {
	if err := kv.Register("consul", New, nil, []string{"endpoint"}); err != nil {
		panic("failed to register consul KV driver")
	}
}
