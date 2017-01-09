package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/urfave/cli.v2"
)

func pgpEncrypt(c *cli.Context, v string) (string, error) {
	var recepients openpgp.EntityList

	to := c.StringSlice("encrypt-for")

	for _, p := range to {
		if p[0] == '/' || p[0] == '.' || p[0] == '~' {
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

	var secretKeyring = os.ExpandEnv(c.String("pgp-sec-ring"))

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
