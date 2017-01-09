package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"golang.org/x/net/context"

	"github.com/nethack42/gokv"

	_ "github.com/nethack42/gokv/providers/etcd"
	_ "github.com/nethack42/gokv/providers/memory"

	"gopkg.in/urfave/cli.v2"
)

func getKV(c *cli.Context) (kv.KV, error) {
	for name, provider := range kv.Providers() {
		if c.Bool(name) {
			params := make(map[string]string)

			for _, key := range provider.RequiredOptions {
				params[key] = c.String(fmt.Sprintf("%s-%s", name, key))
			}

			for _, key := range provider.OptionalOptions {
				params[key] = c.String(fmt.Sprintf("%s-%s", name, key))
			}

			return kv.Open(name, params)
		}
	}

	return nil, fmt.Errorf("no provider specified")
}

var Result interface{}

func main() {
	app := cli.App{}
	app.Name = "gokv"
	app.EnableShellCompletion = true
	app.Version = "0.2.0"
	app.Usage = "A batteries included client to access various Key-Value stores"
	app.Authors = []*cli.Author{
		&cli.Author{
			Name:  "Patrick Pacher",
			Email: "patrick.pacher@gmail.com",
		},
	}

	usr, _ := user.Current()
	dir := usr.HomeDir

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "pgp-sec-ring",
			Aliases: []string{"K"},
			Usage:   "Path to PGP secret keyring used for decryption and signing",
			Value:   dir + "/.gnupg/secring.gpg",
		},
		&cli.StringFlag{
			Name:    "pgp-pub-ring",
			Aliases: []string{"k"},
			Usage:   "Path to PGP public keyring used for encryption and signature verification",
			Value:   dir + "/.gnupg/pubring.gpg",
		},
	}

	app.After = func(c *cli.Context) error {
		if Result != nil {
			return routeOutput(c, Result)
		}
		return nil
	}

	for name, provider := range kv.Providers() {
		flags := []cli.Flag{
			&cli.BoolFlag{
				Name:    name,
				Usage:   fmt.Sprintf("Enable %s Key-Value provider", name),
				EnvVars: []string{"USE_" + strings.ToUpper(name)},
			},
		}

		for _, key := range provider.RequiredOptions {
			flags = append(flags, &cli.StringFlag{
				Name:    fmt.Sprintf("%s-%s", name, key),
				Usage:   fmt.Sprintf("Configure %s for %s provider", key, name),
				EnvVars: []string{fmt.Sprintf("%s_%s", strings.ToUpper(name), strings.ToUpper(key))},
			})
		}

		for _, key := range provider.OptionalOptions {
			flags = append(flags, &cli.StringFlag{
				Name:    fmt.Sprintf("%s-%s", name, key),
				Usage:   fmt.Sprintf("Configure %s for %s provider (optional)", key, name),
				EnvVars: []string{fmt.Sprintf("%s_%s", strings.ToUpper(name), strings.ToUpper(key))},
			})
		}

		app.Flags = append(app.Flags, flags...)
	}

	app.Flags = append(app.Flags, outputFlags()...)

	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "get",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"R"},
					Usage:   "Query recursively (only applies for directories)",
				},
				&cli.BoolFlag{
					Name:    "decrypt",
					Aliases: []string{"d"},
					Usage:   "Try to decrypt the node's value using PGP secret keyring",
				},
			},

			Usage: "Get a key",
			Action: func(c *cli.Context) error {
				k, err := getKV(c)
				if err != nil {
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				var res *kv.Node

				if !c.Bool("recursive") {
					res, err = k.Get(context.Background(), key)
				} else {
					res, err = k.RGet(context.Background(), key)
				}

				if err != nil {
					return err
				}

				if c.Bool("decrypt") {
					decrypted, err := pgpDecrypt(c, string(res.Value))
					if err != nil {
						return err
					}

					res.Value = []byte(decrypted)
				}

				Result = *res

				return nil
			},
		},
		&cli.Command{
			Name:    "delete",
			Usage:   "Delete a key",
			Aliases: []string{"del", "rm"},
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
			Name:    "set",
			Usage:   "Set a key",
			Aliases: []string{"put"},
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "encrypt-for",
					Aliases: []string{"e"},
					Usage:   "Path to PGP public key to encrypt for",
				},
				&cli.BoolFlag{
					Name:    "sign",
					Aliases: []string{"s"},
					Usage:   "Sign value. Only works with --encrypt-for",
				},
				&cli.StringFlag{
					Name:    "file",
					Aliases: []string{"f"},
					Usage:   "Read value from file instead of using commandline parameters. Pass - for stdin",
				},
			},
			Action: func(c *cli.Context) error {
				kv, err := getKV(c)
				if err != nil {
					return err
				}

				key := c.Args().Get(0)
				if key == "" {
					key = "/"
				}

				var value string

				if c.String("file") != "" {
					path := c.String("file")
					var f *os.File

					if path == "-" {
						f = os.Stdin
					} else {
						f_, err := os.Open(path)
						if err != nil {
							return err
						}

						f = f_
					}

					v, err := ioutil.ReadAll(f)
					if err != nil {
						return err
					}

					value = string(v)
				} else {
					value = c.Args().Get(1)
				}

				if len(c.StringSlice("encrypt-for")) > 0 {
					encrypted, err := pgpEncrypt(c, value)
					if err != nil {
						return err
					}

					value = encrypted
				}

				err = kv.Set(context.Background(), key, []byte(value))
				if err != nil {
					return err
				}
				return nil
			},
		},

		&cli.Command{
			Name:    "move",
			Aliases: []string{"mv"},
			Usage:   "Move a key or subtree to a differnt location",
		},
		&cli.Command{
			Name:    "copy",
			Aliases: []string{"cp"},
			Usage:   "Copy a key or subtree to a new location",
		},

		&cli.Command{
			Name:  "proxy",
			Usage: "Launch HTTP API proxy with integrated WebUI",
		},

		&cli.Command{
			Name:    "dump",
			Aliases: []string{"backup"},
			Usage:   "Recursively dump a subtree to file using JSON.",
			Action:  backupTree,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "rel",
					Aliases: []string{"r"},
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
				},
			},
		},
		&cli.Command{
			Name:   "restore",
			Usage:  "Restores a subtree from a JSON file.",
			Action: restoreTree,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "rel",
					Aliases: []string{"r"},
				},
				&cli.StringFlag{
					Name:    "input",
					Aliases: []string{"i"},
				},
			},
		},
	}

	app.Run(os.Args)
}
