package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	from, err := os.Open(c.String("encrypt-for"))
	defer from.Close()

	entityList, err := openpgp.ReadArmoredKeyRing(from)
	if err != nil {
		return "", err
	}

	out := new(bytes.Buffer)

	encOut, err := openpgp.Encrypt(out, entityList, nil, nil, nil)
	if err != nil {
		return "", err
	}

	if _, err := encOut.Write([]byte(v)); err != nil {
		return "", err
	}

	if err := encOut.Close(); err != nil {
		return "", err
	}

	return string(out.Bytes()), nil
}

func pgpDecrypt(c *cli.Context, v string) (string, error) {
	// init some vars
	var entity *openpgp.Entity
	var entityList openpgp.EntityList

	var secretKeyring = c.String("pgp-secret-keyring")
	passphrase, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}

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

	// Get the passphrase and read the private key.
	// Have not touched the encrypted string yet
	passphraseByte := []byte(passphrase)
	log.Println("Decrypting private key using passphrase")
	entity.PrivateKey.Decrypt(passphraseByte)
	for _, subkey := range entity.Subkeys {
		if subkey.PrivateKey.Encrypted {
			subkey.PrivateKey.Decrypt(passphraseByte)
		}
	}
	log.Println("Finished decrypting private key using passphrase")

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes.NewBuffer([]byte(v)), entityList, nil, nil)
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

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "pgp-secret-keyring, k",
			Usage: "Path to PGP secret keyring used for decryption and signing",
		},
		&cli.StringFlag{
			Name:  "pgp-public-keyring, K",
			Usage: "Path to PGP public keyring used for encryption and signature verification",
		},
		&cli.StringFlag{
			Name:  "encrypt-for, e",
			Usage: "Path to PGP public key to encrypt for",
		},
		&cli.BoolFlag{
			Name:  "sign, s",
			Usage: "Sign value. Only works with --encrypt-for",
		},
		&cli.BoolFlag{
			Name:  "decrypt, d",
			Usage: "PGP decrypt value",
		},
	}

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
					Name:  "field, f",
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

				if c.Bool("decrypt") {
					decrypted, err := pgpDecrypt(c, string(res.Value))
					if err != nil {
						return err
					}

					res.Value = []byte(decrypted)
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

				if c.String("encrypt-for") != "" {
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
