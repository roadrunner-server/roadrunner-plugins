- ### GZIP HTTP middleware

```yaml
http:
  address: 127.0.0.1:55555
  max_request_size: 1024
  access_logs: false
  middleware: ["gzip"]

  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

Used to compress incoming or outgoing data with the default gzip compression level.