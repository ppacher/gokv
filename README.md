# gokv

`gokv` is a generic Key-Value library similar to GoLang's `sql`. It provides an
abstraction layer to interact with various Key-Value stores. Due to it's desing,
only basic KV operations (Set, Get, Delete, CompareAndSwap) are supported.

`gokv` is allows to integrate support for various Key-Values stores by just using
this package. It can be used for service discovery, configuration and much more.

For now, only basic Key-Value operations are supported but more enhanced wrappers
for like service discovery or PGP encryption are planned.

## Providers

Currently the following KV providers are supported:

 - etcd
 - temporary in-memory KV map

Support for the following providers is planned:

 - redis
 - consul
 - zookeeper

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
   --pgp-secret-keyring value  Path to PGP secret keyring used for decryption and signing (default: "/home/ppc/.gnupg/secring.gpg")
   --pgp-public-keyring value  Path to PGP public keyring used for encryption and signature verification (default: "/home/ppc/.gnupg/pubring.gpg")
   --memory                    Enable memory Key-Value provider (default: false) [$USE_MEMORY]
   --etcd                      Enable etcd Key-Value provider (default: true) [$USE_ETCD]
   --etcd-endpoints value      Configure endpoints for etcd provider (default: "http://localhost:4001/") [$ETCD_ENDPOINTS]
   --help, -h                  show help (default: false)
   --version, -v               print the version (default: false)
```
