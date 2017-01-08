package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/context"

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
				Name:    name,
				Usage:   fmt.Sprintf("Enable %s Key-Value provider", name),
				EnvVars: []string{"USE_" + strings.ToUpper(name)},
			},
		}

		for _, key := range provider.Keys {
			flags = append(flags, &cli.StringFlag{
				Name:    fmt.Sprintf("%s-%s", name, key),
				Usage:   fmt.Sprintf("Configure %s for %s provider", key, name),
				EnvVars: []string{fmt.Sprintf("%s_%s", strings.ToUpper(name), strings.ToUpper(key))},
			})
		}

		app.Flags = append(app.Flags, flags...)
	}

	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "get",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "field",
					Usage: "Only return the value of field f",
				},
			},

			Usage: "Get a key",
			Action: func(c *cli.Context) error {
				kv, err := getKV(c)
				if err != nil {
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				res, err := kv.Get(context.Background(), key)
				if err != nil {
					return err
				}

				buf := new(bytes.Buffer)
				json.NewEncoder(buf).Encode(res)

				field := c.String("field")
				if field == "" {
					os.Stdout.Write(buf.Bytes())
				} else {
					obj := make(map[string]interface{})

					json.Unmarshal(buf.Bytes(), &obj)

					switch obj[field].(type) {
					case string:
						fmt.Println(obj[field])
					default:
						json.NewEncoder(os.Stdout).Encode(obj[field])
					}
				}

				return nil
			},
		},
		&cli.Command{
			Name:  "delete",
			Usage: "Delete a key",
			Action: func(c *cli.Context) error {
				kv, err := getKV(c)
				if err != nil {
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				err = kv.Delete(context.Background(), key)
				if err != nil {
					return err
				}
				return nil
			},
		},
		&cli.Command{
			Name:  "set",
			Usage: "Set a key",
			Action: func(c *cli.Context) error {
				kv, err := getKV(c)
				if err != nil {
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				value := c.Args().Get(1)

				err = kv.Set(context.Background(), key, []byte(value))
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
