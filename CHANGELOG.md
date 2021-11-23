# CHANGELOG

## v2.6.0 (-.-.2021)

## 👀 New:

- ✏️ **[BETA]** Support for the New Relic observability platform. Sample of the client library might be
  found [here](https://github.com/arku31/roadrunner-newrelic). (Thanks @arku31)   
New Relic middleware is a part of the HTTP plugin, thus configuration should be inside it:

```yaml
http:
  address: 127.0.0.1:15389
  middleware: [ "new_relic" ]
  new_relic:
    app_name: "app"
    license_key: "key"
  pool:
    num_workers: 10
    allocate_timeout: 60s
    destroy_timeout: 60s
```

License key and application name could be set via environment variables: (leave `app_name` and `license_key` empty)

- license_key: `NEW_RELIC_LICENSE_KEY`.
- app_name: `NEW_RELIC_APP_NAME`.

To set the New Relic attributes, the PHP worker should send headers values withing the `rr_newrelic` header key.
Attributes should be separated by the `:`, for example `foo:bar`, where `foo` is a key and `bar` is a value. New Relic
attributes sent from the worker will not appear in the HTTP response, they will be sent directly to the New Relic.

To see the sample of the PHP library, see the @arku31 implementation: https://github.com/arku31/roadrunner-newrelic

The special key which PHP may set to overwrite the transaction name is: `transaction_name`. For
example: `transaction_name:foo` means: set transaction name as `foo`. By default, `RequestURI` is used as the
transaction name.

```php
        $resp = new \Nyholm\Psr7\Response();
        $rrNewRelic = [
            'shopId:1', //custom data
            'auth:password', //custom data
            'transaction_name:test_transaction' //name - special key to override the name. By default it will use requestUri.
        ];

        $resp = $resp->withHeader('rr_newrelic', $rrNewRelic);
```

---

- ✏️ New plugin: `TCP`. The TCP plugin is used to handle raw TCP payload with a bi-directional [protocol](tcp/docs/tcp.md) between the RR server and PHP worker.

PHP client library: https://github.com/spiral/roadrunner-tcp

Configuration:
```yaml
rpc:
  listen: tcp://127.0.0.1:6001

server:
  command: "php ../../psr-worker-tcp-cont.php"

tcp:
  servers:
    server1:
      addr: 127.0.0.1:7778
      delimiter: "\r\n"
    server2:
      addr: 127.0.0.1:8811
      read_buf_size: 10
    server3:
      addr: 127.0.0.1:8812
      delimiter: "\r\n"
      read_buf_size: 1

  pool:
    num_workers: 5
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

---

- ✏️ New HTTP middleware: `http_metrics`. 
```yaml
http:
  address: 127.0.0.1:15389
  middleware: [ "http_metrics" ]
  pool:
    num_workers: 10
    allocate_timeout: 60s
    destroy_timeout: 60s
```
All old and new http metrics will be available after the middleware is activated. Be careful, this middleware may slow down your requests. New metrics:

    - `rr_http_requests_queue_sum` - number of queued requests.
    - `rr_http_no_free_workers_total` - number of the occurrences of the `NoFreeWorkers` errors.  
  

-----
  
- ✏️ New file server to serve static files. It works on a different address, so it doesn't affect the HTTP performance. It uses advanced configuration specific for the static file servers. It can handle any number of directories with its own HTTP prefixes.
Config:

```yaml
fileserver:
  # File server address
  #
  # Error on empty
  address: 127.0.0.1:10101
  # Etag calculation. Request body CRC32.
  #
  # Default: false
  calculate_etag: true

  # Weak etag calculation
  #
  # Default: false
  weak: false

  # Enable body streaming for the files more than 4KB
  #
  # Default: false
  stream_request_body: true

  serve:
    # HTTP prefix
    #
    # Error on empty
  - prefix: "/foo"

    # Directory to serve
    #
    # Default: "."
    root: "../../../tests"

    # When set to true, the server tries minimizing CPU usage by caching compressed files
    #
    # Default: false
    compress: false

    # Expiration duration for inactive file handlers. Units: seconds.
    #
    # Default: 10, use a negative value to disable it.
    cache_duration: 10

    # The value for the Cache-Control HTTP-header. Units: seconds
    #
    # Default: 10 seconds
    max_age: 10

    # Enable range requests
    # https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
    #
    # Default: false
    bytes_range: true

  - prefix: "/foo/bar"
    root: "../../../tests"
    compress: false
    cache_duration: 10s
    max_age: 10
    bytes_range: true
```

- ✏️ `on_init` option for the `server` plugin. `on_init` code executed before the regular command and can be used to warm up the application for example. Failed `on_init` command doesn't affect the main command, so, the RR will continue to run.

Config:
```yaml
# Application server settings (docs: https://roadrunner.dev/docs/php-worker)
server:
  on_init:
    # Command to execute before the main server's command
    #
    # This option is required if using on_init
    command: "any php or script here"
    
    # Script execute timeout
    #
    # Default: 60s [60m, 60h], if used w/o units its means - NANOSECONDS.
    exec_timeout: 20s
    
    # Environment variables for the worker processes.
    #
    # Default: <empty map>
    env:
      - SOME_KEY: "SOME_VALUE"
      - SOME_KEY2: "SOME_VALUE2" 

  # ..REGULAR SERVER OPTIONS...
```

## 🩹 Fixes:

- 🐛 Fix: GRPC server will show message when started.
- 🐛 Fix: Static plugin headers were added to all requests. [BUG](https://github.com/spiral/roadrunner-plugins/issues/115)

## v2.5.2 (27.10.2021)

## 🩹 Fixes:

- 🐛 Fix: panic in the TLS layer. The `http` plugin used `http` server instead of `https` in the rootCA routine.

## v2.5.1 (22.10.2021)

## 🩹 Fixes:

- 🐛 Fix: [base64](https://github.com/spiral/roadrunner-plugins/issues/86) response instead of json in some edge cases.

## v2.5.0 (20.10.2021)

# 💔 Breaking change:

- 🔨 Some drivers now use a new `config` key to handle local configuration. Involved plugins and drivers:
- `plugins`: `broadcast`, `kv`
- `drivers`: `memory`, `redis`, `memcached`, `boltdb`.

### Old style:

```yaml
broadcast:
  default:
    driver: memory
    interval: 1
```

### New style:

```yaml
broadcast:
  default:
    driver: memory
      config: { } <--------------- NEW
```

```yaml
kv:
  memory-rr:
    driver: memory
    config: <--------------- NEW
      interval: 1

kv:
  memcached-rr:
    driver: memcached
    config: <--------------- NEW
      addr:
        - "127.0.0.1:11211"

broadcast:
  default:
    driver: redis
    config: <------------------ NEW
      addrs:
        - "127.0.0.1:6379"
```

## 👀 New:

- ✏️ **[BETA]** GRPC plugin update to v2.
- ✏️ [Roadrunner-plugins](https://github.com/spiral/roadrunner-plugins) repository. This is the new home for the
  roadrunner plugins with documentation, configuration samples, and common problems.
- ✏️ **[BETA]** Let's Encrypt support. RR now can obtain an SSL certificate/PK for your domain automatically. Here is
  the new configuration:

```yaml
    ssl:
      # Host and port to listen on (eg.: `127.0.0.1:443`).
      #
      # Default: ":443"
      address: "127.0.0.1:443"

      # Use ACME certificates provider (Let's encrypt)
      acme:
        # Directory to use as a certificate/pk, account info storage
        #
        # Optional. Default: rr_cache
        certs_dir: rr_le_certs

        # User email
        #
        # Used to create LE account. Mandatory. Error on empty.
        email: you-email-here@email

        # Alternate port for the http challenge. Challenge traffic should be redirected to this port if overridden.
        #
        # Optional. Default: 80
        alt_http_port: 80,


        # Alternate port for the tls-alpn-01 challenge. Challenge traffic should be redirected to this port if overridden.
        #
        # Optional. Default: 443.
        alt_tlsalpn_port: 443,

        # Challenge types
        #
        # Optional. Default: http-01. Possible values: http-01, tlsalpn-01
        challenge_type: http-01

        # Use production or staging endpoints. NOTE, try to use the staging endpoint (`use_production_endpoint`: `false`) to make sure, that everything works correctly.
        #
        # Optional, but for production should be set to true. Default: false
        use_production_endpoint: true

        # List of your domains to obtain certificates
        #
        # Mandatory. Error on empty.
        domains: [
            "your-cool-domain.here",
            "your-second-domain.here"
        ]
```

- ✏️ Add a new option to the `logs` plugin to configure the line ending. By default, used `\n`.

**New option**:

```yaml
# Logs plugin settings
logs:
  (....)
  # Line ending
  #
  # Default: "\n".
  line_ending: "\n"
```

- ✏️ HTTP [Access log support](https://github.com/spiral/roadrunner-plugins/issues/34) at the `Info` log level.

```yaml
http:
  address: 127.0.0.1:55555
  max_request_size: 1024
  access_logs: true <-------- Access Logs ON/OFF
  middleware: [ ]

  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

- ✏️ HTTP middleware to handle `X-Sendfile` [header](https://github.com/spiral/roadrunner-plugins/issues/9). Middleware
  reads the file in 10MB chunks. So, for example for the 5Gb file, only 10MB of RSS will be used. If the file size is
  smaller than 10MB, the middleware fits the buffer to the file size.

```yaml
http:
  address: 127.0.0.1:44444
  max_request_size: 1024
  middleware: [ "sendfile" ] <----- NEW MIDDLEWARE

  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

- ✏️ Service plugin now supports env variables passing to the script/executable/binary/any like in the `server` plugin:

```yaml
service:
  some_service_1:
    command: "php test_files/loop_env.php"
    process_num: 1
    exec_timeout: 5s # s,m,h (seconds, minutes, hours)
    remain_after_exit: true
    env: <----------------- NEW
      foo: "BAR"
    restart_sec: 1
```

- ✏️ Server plugin can accept scripts (sh, bash, etc) in it's `command` configuration key:

```yaml
server:
  command: "./script.sh OR sh script.sh" <--- UPDATED
  relay: "pipes"
  relay_timeout: "20s"
```

The script should start a worker as the last command. For the `pipes`, scripts should not contain programs, which can
close `stdin`, `stdout` or `stderr`.

- ✏️ Nats jobs driver support - [PR](https://github.com/spiral/roadrunner-plugins/pull/68).

```yaml
nats:
  addr: "demo.nats.io"

jobs:
  num_pollers: 10
  pipeline_size: 100000
  pool:
    num_workers: 10
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s

  pipelines:
    test-1:
      driver: nats
      prefetch: 100
      subject: "default"
      stream: "foo"
      deliver_new: "true"
      rate_limit: 100
      delete_stream_on_stop: false
      delete_after_ack: false
      priority: 2

  consume: [ "test-1" ]
```

- Driver uses NATS JetStream API and is not compatible with non-js API.


- ✏️ Response API for the NATS, RabbitMQ, SQS and Beanstalk drivers. This means, that you'll be able to respond to a
  specified in the response queue. Limitations:
    - To send a response to the queue maintained by the RR, you should send it as a `Job` type. There are no limitations
      for the responses into the other queues (tubes, subjects).
    - Driver uses the same endpoint (address) to send the response as specified in the configuration.

## 🩹 Fixes:

- 🐛 Fix: local and global configuration parsing.
- 🐛 Fix: `boltdb-jobs` connection left open after RPC close command.
- 🐛 Fix: close `beanstalk` connection and release associated resources after pipeline stopped.
- 🐛 Fix: grpc plugin fails to handle requests after calling `reset`.
- 🐛 Fix: superfluous response.WriteHeader call when connection is broken.

## 📦 Packages:

- 📦 roadrunner `v2.5.0`
- 📦 roadrunner-plugins `v2.5.0`
- 📦 roadrunner-temporal `v1.0.10`
- 📦 endure `v1.0.6`
- 📦 goridge `v3.2.3`

## v2.4.1 (13.09.2021)

## 🩹 Fixes:

- 🐛 Fix: bug with not-idempotent call to the `attributes.Init`.
- 🐛 Fix: memory jobs driver behavior. Now memory driver starts consuming automatically if the user consumes the
  pipeline in the configuration.

## v2.4.0 (02.09.2021)

## 💔 Internal BC:

- 🔨 Pool, worker interfaces: payload now passed and returned by the pointer.

## 👀 New:

- ✏️ Long-awaited, reworked `Jobs` plugin with pluggable drivers. Now you can allocate/destroy pipelines in the runtime.
  Drivers included in the initial release: `RabbitMQ (0-9-1)`, `SQS v2`, `beanstalk`, `memory` and local queue powered
  by the `boltdb`. [PR](https://github.com/spiral/roadrunner/pull/726)
- ✏️ Support for the IPv6 (`tcp|http(s)|empty [::]:port`, `tcp|http(s)|empty [::1]:port`
  , `tcp|http(s)|empty :// [0:0:0:0:0:0:0:1]:port`) for RPC, HTTP and other
  plugins. [RFC](https://datatracker.ietf.org/doc/html/rfc2732#section-2)
- ✏️ Support for the Docker images via GitHub packages.
- ✏️ Go 1.17 support for the all spiral packages.

## 🩹 Fixes:

- 🐛 Fix: fixed bug with goroutines waiting on the internal worker's container
  channel, [issue](https://github.com/spiral/roadrunner/issues/750).
- 🐛 Fix: RR become unresponsive when new workers failed to
  re-allocate, [issue](https://github.com/spiral/roadrunner/issues/772).
- 🐛 Fix: add `debug` pool config key to the `.rr.yaml`
  configuration [reference](https://github.com/spiral/roadrunner-binary/issues/79).

## 📦 Packages:

- 📦 Update goridge to `v3.2.1`
- 📦 Update temporal to `v1.0.9`
- 📦 Update endure to `v1.0.4`

## 📈 Summary:

- RR Milestone [2.4.0](https://github.com/spiral/roadrunner/milestone/29?closed=1)
- RR-Binary Milestone [2.4.0](https://github.com/spiral/roadrunner-binary/milestone/10?closed=1)

---

## v2.3.2 (14.07.2021)

## 🩹 Fixes:

- 🐛 Fix: Do not call the container's Stop method after the container stopped by an error.
- 🐛 Fix: Bug with ttl incorrectly handled by the worker [PR](https://github.com/spiral/roadrunner/pull/749)
- 🐛 Fix: Add `RR_BROADCAST_PATH` to the `websockets` plugin [PR](https://github.com/spiral/roadrunner/pull/749)

## 📈 Summary:

- RR Milestone [2.3.2](https://github.com/spiral/roadrunner/milestone/31?closed=1)

---

## v2.3.1 (30.06.2021)

## 👀 New:

- ✏️ Rework `broadcast` plugin. Add architecture diagrams to the `doc`
  folder. [PR](https://github.com/spiral/roadrunner/pull/732)
- ✏️ Add `Clear` method to the KV plugin RPC. [PR](https://github.com/spiral/roadrunner/pull/736)

## 🩹 Fixes:

- 🐛 Fix: Bug with channel deadlock when `exec_ttl` was used and TTL limit
  reached [PR](https://github.com/spiral/roadrunner/pull/738)
- 🐛 Fix: Bug with healthcheck endpoint when workers were marked as invalid and stay is that state until next
  request [PR](https://github.com/spiral/roadrunner/pull/738)
- 🐛 Fix: Bugs with `boltdb` storage: [Boom](https://github.com/spiral/roadrunner/issues/717)
  , [Boom](https://github.com/spiral/roadrunner/issues/718), [Boom](https://github.com/spiral/roadrunner/issues/719)
- 🐛 Fix: Bug with incorrect redis initialization and usage [Bug](https://github.com/spiral/roadrunner/issues/720)
- 🐛 Fix: Bug, Goridge duplicate error messages [Bug](https://github.com/spiral/goridge/issues/128)
- 🐛 Fix: Bug, incorrect request `origin` check [Bug](https://github.com/spiral/roadrunner/issues/727)

## 📦 Packages:

- 📦 Update goridge to `v3.1.4`
- 📦 Update temporal to `v1.0.8`

## 📈 Summary:

- RR Milestone [2.3.1](https://github.com/spiral/roadrunner/milestone/30?closed=1)
- Temporal Milestone [1.0.8](https://github.com/temporalio/roadrunner-temporal/milestone/11?closed=1)
- Goridge Milestone [3.1.4](https://github.com/spiral/goridge/milestone/11?closed=1)

---

## v2.3.0 (08.06.2021)

## 👀 New:

- ✏️ Brand new `broadcast` plugin now has the name - `websockets` with broadcast capabilities. It can handle hundreds of
  thousands websocket connections very efficiently (~300k messages per second with 1k connected clients, in-memory bus
  on 2CPU cores and 1GB of RAM) [Issue](https://github.com/spiral/roadrunner/issues/513)
- ✏️ Protobuf binary messages for the `websockets` and `kv` RPC calls under the
  hood. [Issue](https://github.com/spiral/roadrunner/issues/711)
- ✏️ Json-schemas for the config file v1.0 (it also registered
  in [schemastore.org](https://github.com/SchemaStore/schemastore/pull/1614))
- ✏️ `latest` docker image tag supported now (but we strongly recommend using a versioned tag (like `0.2.3`) instead)
- ✏️ Add new option to the `http` config section: `internal_error_code` to override default (500) internal error
  code. [Issue](https://github.com/spiral/roadrunner/issues/659)
- ✏️ Expose HTTP plugin metrics (workers memory, requests count, requests duration)
  . [Issue](https://github.com/spiral/roadrunner/issues/489)
- ✏️ Scan `server.command` and find errors related to the wrong path to a `PHP` file, or `.ph`, `.sh`
  scripts. [Issue](https://github.com/spiral/roadrunner/issues/658)
- ✏️ Support file logger with log rotation [Wiki](https://en.wikipedia.org/wiki/Log_rotation)
  , [Issue](https://github.com/spiral/roadrunner/issues/545)

## 🩹 Fixes:

- 🐛 Fix: Bug with `informer.Workers` worked incorrectly: [Bug](https://github.com/spiral/roadrunner/issues/686)
- 🐛 Fix: Internal error messages will not be shown to the user (except HTTP status code). Error message will be in
  logs: [Bug](https://github.com/spiral/roadrunner/issues/659)
- 🐛 Fix: Error message will be properly shown in the log in case of `SoftJob`
  error: [Bug](https://github.com/spiral/roadrunner/issues/691)
- 🐛 Fix: Wrong applied middlewares for the `fcgi` server leads to the
  NPE: [Bug](https://github.com/spiral/roadrunner/issues/701)

## 📦 Packages:

- 📦 Update goridge to `v3.1.0`

---

## v2.2.1 (13.05.2021)

## 🩹 Fixes:

- 🐛 Fix: revert static plugin. It stays as a separate plugin on the main route (`/`) and supports all the previously
  announced features.
- 🐛 Fix: remove `build` and other old targets from the Makefile.

---

## v2.2.0 (11.05.2021)

## 👀 New:

- ✏️ Reworked `static` plugin. Now, it does not affect the performance of the main route and persist on the separate
  file server (within the `http` plugin). Looong awaited feature: `Etag` (+ weak Etags) as well with the `If-Mach`
  , `If-None-Match`, `If-Range`, `Last-Modified`
  and `If-Modified-Since` tags supported. Static plugin has a bunch of new options such as: `allow`, `calculate_etag`
  , `weak` and `pattern`.

  ### Option `always` was deleted from the plugin.

- ✏️ Update `informer.List` implementation. Now it returns a list with the all available plugins in the runtime.

## 🩹 Fixes:

- 🐛 Fix: issue with wrong ordered middlewares (reverse). Now the order is correct.
- 🐛 Fix: issue when RR fails if a user sets `debug` mode with the `exec_ttl` supervisor option.
- 🐛 Fix: uniform log levels. Use everywhere the same levels (warn, error, debug, info, panic).

---

## v2.1.1 (29.04.2021)

## 🩹 Fixes:

- 🐛 Fix: issue with endure provided wrong logger interface implementation.

## v2.1.0 (27.04.2021)

## 👀 New:

- ✏️ New `service` plugin. Docs: [link](https://roadrunner.dev/docs/beep-beep-service)
- ✏️ Stabilize `kv` plugin with `boltdb`, `in-memory`, `memcached` and `redis` drivers.

## 🩹 Fixes:

- 🐛 Fix: Logger didn't provide an anonymous log instance to a plugins w/o `Named` interface implemented.
- 🐛 Fix: http handler was without log listener after `rr reset`.

## v2.0.4 (06.04.2021)

## 👀 New:

- ✏️ Add support for `linux/arm64` platform for docker image (thanks @tarampampam).
- ✏️ Add dotenv file support (`.env` in working directory by default; file location can be changed using CLI
  flag `--dotenv` or `DOTENV_PATH` environment variable) (thanks @tarampampam).
- 📜 Add a new `raw` mode for the `logger` plugin to keep the stderr log message of the worker unmodified (logger
  severity level should be at least `INFO`).
- 🆕 Add Readiness probe check. The `status` plugin provides `/ready` endpoint which return the `204` HTTP code if there
  are no workers in the `Ready` state and `200 OK` status if there are at least 1 worker in the `Ready` state.

## 🩹 Fixes:

- 🐛 Fix: bug with the temporal worker which does not follow general graceful shutdown period.

## v2.0.3 (29.03.2021)

## 🩹 Fixes:

- 🐛 Fix: slow last response when reached `max_jobs` limit.

## v2.0.2 (06.04.2021)

- 🐛 Fix: Bug with required Root CA certificate for the SSL, now it's optional.
- 🐛 Fix: Bug with incorrectly consuming metrics collector from the RPC calls (thanks @dstrop).
- 🆕 New: HTTP/FCGI/HTTPS internal logs instead of going to the raw stdout will be displayed in the RR logger at
  the `Info` log level.
- ⚡ New: Builds for the Mac with the M1 processor (arm64).
- 👷 Rework ServeHTTP handler logic. Use http.Error instead of writing code directly to the response writer. Other small
  improvements.

## v2.0.1 (09.03.2021)

- 🐛 Fix: incorrect PHP command validation
- 🐛 Fix: ldflags properly inject RR version
- ⬆️ Update: README, links to the go.pkg from v1 to v2
- 📦 Bump golang version in the Dockerfile and in the `go.mod` to 1.16
- 📦 Bump Endure container to v1.0.0.

## v2.0.0 (02.03.2021)

- ✔️ Add a shared server to create PHP worker pools instead of isolated worker pool in each individual plugin.
- 🆕 New plugin system with auto-recovery, easier plugin API.
- 📜 New `logger` plugin to configure logging for each plugin individually.
- 🔝 Up to 50% performance increase in HTTP workloads.
- ✔️ Add **[Temporal Workflow](https://temporal.io)** plugin to run distributed computations on scale.
- ✔️ Add `debug` flag to reload PHP worker ahead of a request (emulates PHP-FPM behavior).
- ❌ Eliminate `limit` service, now each worker pool includes `supervisor` configuration.
- 🆕 New resetter, informer plugins to perform hot reloads and observe loggers in a system.
- 💫 Expose more HTTP plugin configuration options.
- 🆕 Headers, static and gzip services now located in HTTP config.
- 🆕 Ability to configure the middleware sequence.
- 💣 Faster Goridge protocol (eliminated 50% of syscalls).
- 💾 Add support for binary payloads for RPC (`msgpack`).
- 🆕 Server no longer stops when a PHP worker dies (attempts to restart).
- 💾 New RR binary server downloader.
- 💣 Echoing no longer breaks execution (yay!).
- 🆕 Migration to ZapLogger instead of Logrus.
- 💥 RR can no longer stuck when studding down with broken tasks in a pipeline.
- 🧪 More tests, more static analysis.
- 💥 Create a new foundation for new KV, WebSocket, GRPC and Queue plugins.

## v2.0.0-RC.4 (20.02.2021)

- PHP tests use latest signatures (https://github.com/spiral/roadrunner/pull/550).
- Endure container update to v1.0.0-RC.2 version.
- Remove unneeded mutex from the `http.Workers` method.
- Rename `checker` plugin package to `status`, remove `/v1` endpoint prefix (#557).
- Add static, headers, status, gzip plugins to the `main.go`.
- Fix workers pool behavior -> idle_ttl, ttl, max_memory are soft errors and exec_ttl is hard error.

## v2.0.0-RC.3 (17.02.2021)

- Add support for the overwriting `.rr.yaml` keys with values (ref: https://roadrunner.dev/docs/intro-config)
- Make logger plugin optional to define in the config. Default values: level -> `debug`, mode -> `development`
- Add the ability to read env variables from the `.rr.yaml` in the form of: `rpc.listen: {RPC_ADDR}`. Reference:
  ref: https://roadrunner.dev/docs/intro-config (Environment Variables paragraph)

## v2.0.0-RC.2 (11.02.2021)

- Update RR to version v2.0.0-RC.2
- Update Temporal plugin to version v2.0.0-RC.1
- Update Goridge to version v3.0.1
- Update Endure to version v1.0.0-RC.1
