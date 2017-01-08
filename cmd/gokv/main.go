package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/nethack42/gokv"

	_ "github.com/nethack42/gokv/providers/etcd"
	_ "github.com/nethack42/gokv/providers/memory"

	"gopkg.in/urfave/cli.v2"
)

func getKV(c *cli.Context) (kv.KV, error) {
	for name, provider := range kv.Factories() {
		if c.Bool(name) {
			params := make(map[string]string)

			for _, key := range provider.Keys {
				params[key] = c.String(fmt.Sprintf("%s-%s", name, key))
			}

			return kv.Open(name, params)
		}
	}

	return nil, fmt.Errorf("no provider specified")
}

func main() {
	app := cli.App{}
	app.Name = "gokv"
	app.Usage = "GoKV is a generic Key-Value client"

	app.Flags = []cli.Flag{}

	for name, provider := range kv.Factories() {
		flags := []cli.Flag{
			&cli.BoolFlag{
				Name:  name,
				Usage: fmt.Sprintf("Enable %s Key-Value provider", name),
			},
		}

		for _, key := range provider.Keys {
			flags = append(flags, &cli.StringFlag{
				Name:  fmt.Sprintf("%s-%s", name, key),
				Usage: fmt.Sprintf("Configure %s for %s provider", key, name),
			})
		}

		app.Flags = append(app.Flags, flags...)
	}

	app.Commands = []*cli.Command{
		&cli.Command{
			Name:  "get",
			Usage: "Get a key",
			Action: func(c *cli.Context) error {
				kv, err := getKV(c)
				if err != nil {
					logrus.Error(err)
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				res, err := kv.Get(context.Background(), key)
				if err != nil {
					logrus.Error(err)
					return err
				}

				json.NewEncoder(os.Stdout).Encode(res)

				return nil
			},
		},
	}

	app.Run(os.Args)
}
