# gokv

**gokv** provides a generic access layer for various Key-Value databases (including
**etcd**, **consul**, ...) as well as a batteries-included command line client with support for recursive dumps/backups, PGP encryption/signatures, interactive mode (*coming soon*) and more. Btw, adding support for new KV databases is easy.

**gokv** is meant to be used for configuration backends or service discovery 
regardless of the underlying KV provider. It is designed similar to Golang's
`database/sql`. It is both, a [library](#Library-Usage) as 
well as a command line [cli](#Commandline-Client).

At the moment, the following backend providers are supported:

- **[etcd](providers/etcd/README.md)**
- **memory** *temporary in-memory KV storage*
- **consul** *partial; work-in-progress*

In addition, support for **ZooKeeper**, **Redis**, **Memcache**, **cznic/kv**,
**bolt** and **tiedot** is planned.

**Note**: gokv is still under heavy development and until we reach a final 1.0.0
APIs may change with any 0.x release.

## Library Usage

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

    val, _ := store.RGet(ctx, "/a/b") // Recursively query /a/b

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
   gokv - A batteries included client to access various Key-Value stores

USAGE:
   gokv [global options] command [command options] [arguments...]

VERSION:
   0.2.0

AUTHOR:
   Patrick Pacher <patrick.pacher@gmail.com>

COMMANDS:
     get              Get a key
     delete, del, rm  Delete a key
     set, put         Set a key
     move, mv         Move a key or subtree to a differnt location
     copy, cp         Copy a key or subtree to a new location
     proxy            Launch HTTP API proxy with integrated WebUI
     dump             Recursively dump a subtree to file using JSON.
     restore          Restores a subtree from a JSON file.
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --pgp-sec-ring value, -K value  Path to PGP secret keyring used for decryption and signing (default: "~/.gnupg/secring.gpg")
   --pgp-pub-ring value, -k value  Path to PGP public keyring used for encryption and signature verification (default: "~/.gnupg/pubring.gpg")
   --etcd                          Enable etcd Key-Value provider (default: true) [$USE_ETCD]
   --etcd-endpoints value          Configure endpoints for etcd provider (default: "http://localhost:4001/") [$ETCD_ENDPOINTS]
   --memory                        Enable memory Key-Value provider (default: false) [$USE_MEMORY]
   --json, -j                      Display result as JSON (default: false)
   --list, -l                      Only display the nodes children as a list (default: false)
   --tree, -t                      Only display the nodes children as a tree (default: false)
   --value, -o                     Only display the nodes value (default: false)
   --help, -h                      show help (default: false)
   --version, -v                   print the version (default: false)
```

### Shell Completion

You can generate bash or zsh completion code by using the flag `--init-completion bash` or `--init-completion zsh`.

To setup for bash:

```bash
eval "`gokv --init-completion bash`"
```

Alternatively, you can put the completion code in your `.bashrc` file:
```bash
gokv --init-completion bash >> ~/.bashrc
```

### Advanced Usage

The `gokv` command line client provides some handy features for more complex
operations including PGP encryption, sub-tree backup & restore and much more.

#### Backup & Restore

Using `gokv backup` and `gokv restore` it is possible to dump and restore complete
subtrees. 

```bash
$ gokv dump /coredns/zones/example.com > /tmp/example.com.zone
```

```bash
$ gokv restore -f /tmp/example.com.zone /coredns/zones/example.com
```

Like when using `gokv move` or `gokv copy` it is also possible to move/copy subtrees using backup&restore.
This comes in handy if you want to move a subtree from one KV provider to another:

```bash
# Define endpoints within the environment to keep gokv params simple 
export ETCD_ENDPOINTS="https://etcdnode1:4001"
export CONSUL_ENDPOINTS="https://consul:2000/"

# Copy /config from etcd to consul
gokv --etcd dump /config | gokv --consul restore /app1

# Remove /config from etcd
gokv --etcd rm /config
```

By default, `gokv` operates in `--rel` relative mode. In the example shown above,
the `dump` command runs with `--rel` enabled and thus, removes the prefix /config
from all nodes. In addition, `restore` will also run in relative mode and prefix
each key with /app1.

```
/config                 ->          /app1
    |-main.conf         ->             |-main.conf
    |-example.conf      ->             |-example.conf
```

You can disable relative mode by passing `--rel=false`. Disabling will cause 
`dump` to not modify keys and `restore` to not append any prefix.

#### Using PGP

The `gokv` cli includes basic PGP support. En/Decryption works but siging/verification
is not yet implemented. In addition, keyring support is hacky..

Encrypting your credit-card number for Bob and Frank and store it under /alice/credit-card

```bash
$ gokv set --encrypt-for /path/to/bob.pubkey --encrypt-for /path/to/frank.pubkey /alice/credit-card "XXXX-XXXX-XXXX-XXXX"
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

Instead of letting gokv decrypt the value, you can also rely on `gpg` if more
advanced options are required.

```bash
# Use gpg for encryption
echo "Some-Message" | gpg -e -r patrick.pacher@gmail.com | base64 | gokv set -f - /key/path

# Use gpg for decryption
gokv get --value /alice/credit-card | base64 -d | gpg -d
```

Note: `gokv` applies base64 encoding to encrypted values.

# Contributing

I will gladly accept Pull-Requests for new providers and bug-fixes! If you are going to modify some of the core APIs, please make sure to also update **all** supported providers so tests won't start failing. Also, good code has tests, so please submit some with your PR! We'll try to keep the average line coverage above 80%. Use `go test -covermode=count` to get line-coverage for each package. 

# Roadmap

The list below is a short summary of the projects roadmap. More information can
be found in [Issues](https://github.com/nethack42/gokv/issues) and 
[Milestones](https://github.com/nethack42/gokv/milestone).

**v0.3** (*next*)
 - [ ] Proxy and WebUI support
 - [ ] PGP agent support
 - [ ] Advanced/better error handling in `gokv` cli
 - [ ] New provider: `redis`
 - [ ] Interactive mode (readline support in v0.4)
 - [ ] Clipboard support
 - [ ] Extended passphrase support (env, parameter, external-commands)

**v0.2** (**active**)
 - [X] Support for recursive gets
 - [X] Output types for `gokv` cli: JSON, Value, List, Tree
 - [ ] New provider: `consul`
 - [ ] PGP Keyring support
 - [X] PGP Multi-Receipient encryption
 - [ ] PGP Signature
 - [X] Backup Command
 - [X] Restore Command
 - [ ] Copy and Move commands
 - [X] Shell Completion (zsh, bash) *thanks to urfave/cli*

# Changelog

**v0.1** (released 2017-01-08)
 - Basic PGP support (encrypt/decrypt)
 - Basic KV operations (Get/Set/Delete)
 - New Provider: `etcd`
 - New Provider: `memory`
