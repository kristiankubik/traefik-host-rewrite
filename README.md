# traefik-host-rewrite

A Traefik middleware plugin for rewriting the HTTP request `Host` header.

It supports replacing either an exact hostname or a DNS suffix while preserving any preceding subdomains.

The middleware modifies only the HTTP request `Host` header. It does **not** modify DNS resolution, TLS SNI, or the request URL path.

## Installation

Add the plugin to the static Traefik configuration (traefik.yml):

```yaml
experimental:
    plugins:
        hostrewrite:
            moduleName: github.com/kristiankubik/traefik-host-rewrite
            version: v0.1.0
```

**Restart Traefik after enabling the plugin.**

The plugin can then be used as a middleware in the dynamic configuration:

```yaml
http:
    middlewares:
        rewrite-host:
            plugin:
                hostrewrite:
                    source: example.com
                    target: example.internal
```

Attach the middleware to a router:

```yaml
http:
    routers:
        example:
            rule: "Host(`example.com`)"
            middlewares:
                - rewrite-host
            service: example
```

The plugin rewrites the request `Host` before it is forwarded to the configured service.

## Configuration

```yaml
http:
    middlewares:
        rewrite-host:
            plugin:
                hostrewrite:
                    source: example.com
                    target: example.internal
                    mode: suffix
                    debug: false
```

### Options

| Option   | Required | Default  | Description                               |
| -------- | -------- | -------- | ----------------------------------------- |
| `source` | yes      |          | Hostname or DNS suffix to match           |
| `target` | yes      |          | Hostname or DNS suffix to replace it with |
| `mode`   | no       | `suffix` | Rewrite mode: `suffix` or `replace`       |
| `debug`  | no       | `false`  | Log successful host rewrites              |

## Rewrite modes

### `suffix`

Replaces the matching DNS suffix while preserving preceding subdomains.

```text
source: example.com
target: example.internal

example.com          → example.internal
foo.example.com      → foo.example.internal
a.b.example.com      → a.b.example.internal
other.example.net    → unchanged
```

### `replace`

Replaces the host only when it **exactly matches** `source`.

```text
source: foo.example.com
target: backend.internal

foo.example.com      → backend.internal
x.foo.example.com    → unchanged
other.example.com    → unchanged
```

## Debug logging

Debug logging can be enabled per middleware:

```yaml
http:
    middlewares:
        rewrite-host:
            plugin:
                hostrewrite:
                    source: example.com
                    target: example.internal
                    debug: true
```

A **successful rewrite** will create a log entry similar to:

```text
hostrewrite: mode="suffix" source="example.com" target="example.internal" host="foo.example.com" -> "foo.example.internal"
```
