# gokv

**gokv** provides a generic access layer for various Key-Value databases including
**etcd**, **consul** and many more. Adding support for new KV databases is easy.

**gokv** is meant to be used for configuration backends or service discovery 
regardless of the underlying KV provider. It is designed similar to Golang's
`database/sql`. 

At the moment, the following backend providers are supported:

- **[etcd](providers/etcd/README.md)**
- **memory** *temporary in-memory KV storage*
- **consul** *partial; work-in-progress*

In addition, support for **ZooKeeper**, **Redis**, **Memcache**, **cznic/kv**,
**bolt** and **tiedot** is planned.

## Usage

```golang
package main

import "fmt"
import "context"
import "github.com/nethack42/gokv"
import _ "github.com/nethack42/gokv/providers/etcd"

func main() {
    store, _ := kv.Open("etcd", map[string]string{
        "endpoints": "http://localhost:4001",
    })

    // store, _ := kv.Open("memory", nil)

    ctx := context.Background()

    store.Set(ctx, "/a/b/c", "some-value")

    val, _ := store.Get(ctx, "/a/b")

    for _, child := range val.Children {
        prefix := "f"
        if child.IsDir {
            prefix = "d" 
        }

        fmt.Println(" - [%s] %s %s", prefix, key, child.Value)
    }

    store.Delete(ctx, "/a")
}
```

## Commandline Client

`gokv` also ships a command line client in `cmd/gokv`. In order to install it,
issue the following command:

```bash
$ go get github.com/nethack42/gokv
$ go install github.com/nethack42/gokv/cmd/gokv
```

### Usage

```bash
$ gokv --help
NAME:
gokv - GoKV is a generic Key-Value client

USAGE:
gokv [global options] command [command options] [arguments...]

VERSION:
0.0.0

COMMANDS:
get      Get a key
delete   Delete a key
set      Set a key
help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
--pgp-secret-keyring value  Path to PGP secret keyring used for decryption and signing (default: "/home/alice/.gnupg/secring.gpg")
--pgp-public-keyring value  Path to PGP public keyring used for encryption and signature verification (default: "/home/alice/.gnupg/pubring.gpg")
--memory                    Enable memory Key-Value provider (default: false) [$USE_MEMORY]
--etcd                      Enable etcd Key-Value provider (default: true) [$USE_ETCD]
--etcd-endpoints value      Configure endpoints for etcd provider (default: "http://localhost:4001/") [$ETCD_ENDPOINTS]
--help, -h                  show help (default: false)
--version, -v               print the version (default: false)
```

#### PGP Examples

The `gokv` cli includes basic PGP support. En/Decryption works but siging/verification
is not yet implemented. In addition, keyring support is hacky..

Encrypting your credit-card number for Bob and store it under /alice/credit-card

```bash
$ gokv set --encrypt-for /path/to/bob_pub.key /alice/credit-card "XXXX-XXXX-XXXX-XXXX"
```

Now, as Bob, I can get Alice credit-card number by issuing the following command:

```bash
$ gokv get --json --decrypt /alice/credit-card
{
    "key": "/alice/credit-card",
    "value": "XXXX-XXXX-XXXX-XXXX"
}
```

If your keyring is password protected, `gokv` will ask you on the terminal.
