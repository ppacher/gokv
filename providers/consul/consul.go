package consul

import (
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

func New(params map[string]string) (kv.KV, error) {
	return &KV{}, nil
}

func init() {
	kv.Register("consul", New)
}
