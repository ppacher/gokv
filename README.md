# gokv

`gokv` is a generic Key-Value library similar to GoLang's `sql`. It provides an
abstraction layer to interact with various Key-Value stores. Due to it's desing,
only basic KV operations (Set, Get, Delete, CompareAndSwap) are supported.

`gokv` is allows to integrate support for various Key-Values stores by just using
this package. It can be used for service discovery, configuration and much more.

Currently the following KV providers are supported:

 - etcd
 - temporary in-memory KV map

Support for the following providers is planned:

 - redis
 - consul
 - zookeeper
