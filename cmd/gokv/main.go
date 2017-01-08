package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"

	"github.com/nethack42/gokv"

	_ "github.com/nethack42/gokv/providers/etcd"
	_ "github.com/nethack42/gokv/providers/memory"

	"gopkg.in/urfave/cli.v2"
)

func pgpEncrypt(c *cli.Context, v string) (string, error) {
	//to, err := os.Open(os.ExpandEnv(c.String("encrypt-for")))
	//defer to.Close()

	var recepients openpgp.EntityList

	to := c.StringSlice("encrypt-for")

	for _, p := range to {
		if p[0] == '/' || p[0] == '.' {
			f, err := os.Open(os.ExpandEnv(p))
			defer f.Close()
			// file path
			r, err := openpgp.ReadArmoredKeyRing(f)
			if err != nil {
				return "", err
			}

			recepients = append(recepients, r...)
		}
		// TODO: add support to load from public keyring
	}

	out := new(bytes.Buffer)

	encOut, err := openpgp.Encrypt(out, recepients, nil, nil, nil)
	if err != nil {
		return "", err
	}

	if _, err := encOut.Write([]byte(v)); err != nil {
		return "", err
	}

	if err := encOut.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func pgpDecrypt(c *cli.Context, v string) (string, error) {
	// init some vars
	var entity *openpgp.Entity
	var entityList openpgp.EntityList

	var secretKeyring = os.ExpandEnv(c.String("pgp-secret-keyring"))

	// Open the private key file
	keyringFileBuffer, err := os.Open(secretKeyring)
	if err != nil {
		return "", err
	}
	defer keyringFileBuffer.Close()
	entityList, err = openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return "", err

	}
	entity = entityList[0]

	if entity.PrivateKey.Encrypted {
		fmt.Println("Enter PGP Keyring passphrase: ")
		passphrase, err := terminal.ReadPassword(0)
		if err != nil {
			return "", err
		}

		// Get the passphrase and read the private key.
		// Have not touched the encrypted string yet
		passphraseByte := []byte(passphrase)

		entity.PrivateKey.Decrypt(passphraseByte)
		for _, subkey := range entity.Subkeys {
			if subkey.PrivateKey.Encrypted {
				subkey.PrivateKey.Decrypt(passphraseByte)
			}
		}
	}

	dec, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", err
	}

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes.NewBuffer(dec), entityList, nil, nil)
	if err != nil {
		return "", err

	}
	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err

	}
	decStr := string(bytes)

	return decStr, nil
}

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

func printNode(c *cli.Context, n kv.Node) {
	value := c.Bool("value")

	if value {
		fmt.Println(string(n.Value))
		return
	}

	if !c.Bool("json") {
		fmt.Println("key: ", n.Key)

		fmt.Println("value: ", string(n.Value))

		if n.IsDir {
			childStr := ""

			for _, c := range n.Children {
				childStr += ", " + c.Key
			}
			childStr = childStr[0:]
			fmt.Println("childs: ", childStr)
		} else {
			fmt.Println("childs: ")
		}

		fmt.Println("dir: ", n.IsDir)

		for _, child := range n.Children {
			fmt.Println("")
			printNode(c, child)
		}

	} else {
		json.NewEncoder(os.Stdout).Encode(struct {
			kv.Node
			Value string `json:"value, omitempty"`
		}{
			Node:  n,
			Value: string(n.Value),
		})
	}
}

func main() {
	app := cli.App{}
	app.Name = "gokv"
	app.Usage = "GoKV is a generic Key-Value client"

	usr, _ := user.Current()
	dir := usr.HomeDir

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "pgp-secret-keyring, k",
			Usage: "Path to PGP secret keyring used for decryption and signing",
			Value: dir + "/.gnupg/secring.gpg",
		},
		&cli.StringFlag{
			Name:  "pgp-public-keyring, K",
			Usage: "Path to PGP public keyring used for encryption and signature verification",
			Value: dir + "/.gnupg/pubring.gpg",
		},
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

	app.Commands = []*cli.Command{
		&cli.Command{
			Name: "get",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "key",
					Usage: "Show result key",
				},
				&cli.BoolFlag{
					Name:  "value",
					Usage: "Only print the result value",
				},

				/* TODO
				&cli.BoolFlag{
					Name:  "recursive",
					Usage: "Query recursively (only applies for directories)",
				},
				*/
				&cli.BoolFlag{
					Name:  "json",
					Usage: "Display result as JSON",
				},
				&cli.BoolFlag{
					Name:  "decrypt",
					Usage: "PGP decrypt value",
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

				if c.Bool("decrypt") {
					decrypted, err := pgpDecrypt(c, string(res.Value))
					if err != nil {
						return err
					}

					res.Value = []byte(decrypted)
				}

				printNode(c, *res)

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
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:  "encrypt-for, e",
					Usage: "Path to PGP public key to encrypt for",
				},
				&cli.BoolFlag{
					Name:  "sign, s",
					Usage: "Sign value. Only works with --encrypt-for",
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

				value := c.Args().Get(1)

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
	}

	app.Run(os.Args)
}
