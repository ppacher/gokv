package etcd

import (
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/nethack42/gokv"
	"golang.org/x/net/context"
)

type KV struct {
	cli   client.Client
	store client.KeysAPI
}

func (e *KV) Set(ctx context.Context, key string, value []byte) error {
	key = sanatizePath(key)
	_, err := e.store.Set(ctx, key, string(value), nil)

	return err
}

func (e *KV) Get(ctx context.Context, key string) (*kv.Node, error) {
	key = sanatizePath(key)
	node, err := e.store.Get(ctx, key, &client.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}

	res := convertNode(node.Node)

	return res, nil
}

func convertNode(n *client.Node) *kv.Node {
	node := &kv.Node{
		Key:   sanatizePath(n.Key),
		IsDir: n.Dir,
	}

	if n.Dir {
		for _, child := range n.Nodes {
			node.Children = append(node.Children, *convertNode(child))
		}
	} else {
		if n.Value != "" {
			node.Value = []byte(n.Value)
		}
	}
	return node
}

func sanatizePath(path string) string {
	return strings.Trim(path, "/")
}

func (e *KV) Delete(ctx context.Context, key string) error {
	key = sanatizePath(key)

	node, err := e.Get(ctx, key)
	if err != nil {
		return err
	}

	_, err = e.store.Delete(ctx, key, &client.DeleteOptions{
		Dir:       node.IsDir,
		Recursive: true,
	})
	return err
}

func (e *KV) CAS(ctx context.Context, key string, compar, value []byte) error {
	return nil
}

func New(params map[string]string) (kv.KV, error) {
	cli, err := client.New(client.Config{
		Endpoints:               strings.Split(params["endpoints"], ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})

	if err != nil {
		return nil, err
	}

	e := &KV{
		cli:   cli,
		store: client.NewKeysAPI(cli),
	}

	return e, nil
}

func init() {
	kv.Register("etcd", New, []string{"endpoints"}, nil)
}
