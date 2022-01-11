## Cache (RFC7234) middleware [WIP]

Cache middleware implements http-caching RFC 7234 (not fully yet).  
It handles the following headers:

- `Cache-Control`
- `max-age`

**Responses**:

- `Age`

**HTTP codes**:

- `OK (200)`

**Available backends**:

- `memory`

**Available methods**:

- `GET`

## Configuration

```yaml
http:
  address: 127.0.0.1:44933
  middleware: ["cache"]
  # ...
  cache:
    driver: memory
    cache_methods: ["GET", "HEAD", "POST"] # only GET by default
    config: {}
```

For the worker sample and other docs, please, refer to the [http plugin](http.md)