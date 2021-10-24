- ### X-Sendfile

```yaml
http:
  address: 127.0.0.1:55555
  max_request_size: 1024
  access_logs: false
  middleware: ["sendfile"]

  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

HTTP middleware to handle `X-Sendfile` [header](https://github.com/spiral/roadrunner-plugins/issues/9)
Middleware reads the file in 10MB chunks. So, for example for the 5Gb file, only 10MB of RSS will be used. If the file size is smaller than 10MB, the middleware fits the buffer to the file size.
