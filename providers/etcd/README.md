# `etcd` Provider

This package contains the `etcd` provider for gokv.

To use this provide in your project, include the following line in `main.go` (or similar):


```golang
import _ "github.com/nethack42/gokv/providers/etcd"
```

```golang
store, _ := kv.Open("etcd", map[string]string{
    "endpoints": "http://node1:4001,http://node2:4001",
})
```

## Parameters

### `endpoints`

**Required**

Should contain one or more comma separated etcd endpoint URLs.

