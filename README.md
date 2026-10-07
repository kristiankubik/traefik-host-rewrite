# traefik-host-rewrite

The middleware modifies the HTTP request Host header. It does not modify DNS resolution, TLS SNI, or the request URL path.

```yaml
http:
    middlewares:
        rewrite-host:
            plugin:
                hostrewrite:
                    source: oni.example.com
                    target: oni.internal
                    debug: true|false (optional)
```
