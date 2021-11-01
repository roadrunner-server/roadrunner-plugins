# roadrunner-plugins

<p align="center">
 <img src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png" height="75px" alt="RoadRunner">
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://pkg.go.dev/github.com/spiral/roadrunner/v2?tab=doc"><img src="https://godoc.org/github.com/spiral/roadrunner/v2?status.svg"></a>
	<a href="https://github.com/spiral/roadrunner/actions"><img src="https://github.com/spiral/roadrunner/workflows/Linux/badge.svg" alt=""></a>
	<a href="https://github.com/spiral/roadrunner/actions"><img src="https://github.com/spiral/roadrunner/workflows/Linters/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/spiral/roadrunner"><img src="https://goreportcard.com/badge/github.com/spiral/roadrunner"></a>
	<a href="https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master"><img src="https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png"></a>
	<a href="https://codecov.io/gh/spiral/roadrunner/"><img src="https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://lgtm.com/projects/g/spiral/roadrunner/alerts/"><img alt="Total alerts" src="https://img.shields.io/lgtm/alerts/g/spiral/roadrunner.svg?logo=lgtm&logoWidth=18"/></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
	<a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://img.shields.io/packagist/dd/spiral/roadrunner?style=flat-square"></a>
	<a href="https://pkg.go.dev/github.com/spiral/roadrunner/v2?tab=doc"><img src="https://godoc.org/github.com/spiral/roadrunner/v2?status.svg"></a>
</p>

Home for the roadrunner plugins.

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, load balancer, and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis.

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup
with much greater performance and flexibility.

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> |
	<a href="https://roadrunner.dev/docs"><b>Documentation</b></a> |
    <a href="https://github.com/orgs/spiral/projects/2"><b>Release schedule</b></a>
</p>

# Available Plugins

| Plugin                                              | Description                                                                                                                         | Docs                          |
| --------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| ![](https://img.shields.io/badge/-HTTP-green)       | Provides HTTP, HTTPS, FCGI transports. Extensible with middleware.                                                                  | [Docs](http/docs/http.md)
| ![](https://img.shields.io/badge/-Headers-blue)     | HTTP middleware supports constant custom headers and CORS.                                                                          | [Docs](http/docs/headers.md)
| ![](https://img.shields.io/badge/-GZIP-blue)        | HTTP middleware supports `Accept-Encoding`: GZIP.                                                                                   | [Docs](http/docs/gzip.md)
| ![](https://img.shields.io/badge/-Static-blue)      | HTTP middleware serves static files.                                                                                                | [Docs](http/docs/static.md)
| ![](https://img.shields.io/badge/-Sendfile-blue)    | HTTP middleware handles X-Sendfile headers.                                                                                         | [Docs](http/docs/sendfile.md)
| ![](https://img.shields.io/badge/-NewRelic-blue)    | HTTP middleware supports NewRelic distributed traces and custom attributes.                                                         | [Docs](http/docs/new_relic.md)

| Plugin                                              | Description                                                                                                                         | Docs                          |
| --------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| ![](https://img.shields.io/badge/-Jobs-green)       | Provides queues support for the RR2 via different drivers                                                                           | [Docs](jobs/docs/jobs.md)
| ![](https://img.shields.io/badge/-AMQP-blue)        | Provides AMQP (0-9-1) protocol support via RabbitMQ                                                                                 | [Docs](amqp/docs/amqp_jobs.md)
| ![](https://img.shields.io/badge/-Beanstalk-blue)   | Provides [beanstalkd](https://github.com/beanstalkd/beanstalkd) queue support                                                       | [Docs](beanstalk/docs/beanstalk_jobs.md)
| ![](https://img.shields.io/badge/-Boltdb-blue)      | Provides support for the [BoltDB](https://github.com/etcd-io/bbolt) key/value store. Used in the `Jobs` and `KV`                    | [Docs](boltdb/docs/boltdb_jobs.md)
| ![](https://img.shields.io/badge/-SQS-blue)         | SQS driver for the jobs                                                                                                             | [Docs](sqs/docs/sqs_jobs.md)
| ![](https://img.shields.io/badge/-NATS-blue)        | NATS jobs driver                                                                                                                    | [Docs](nats/docs/nats.md)     |

| Plugin                                              | Description                                                                                                                         | Docs                          |
| --------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| ![](https://img.shields.io/badge/-KV-green)         | Provides key-value support for the RR2 via different drivers                                                                        |
| ![](https://img.shields.io/badge/-Memcached-blue)   | Memcached driver for the kv                                                                                                         |
| ![](https://img.shields.io/badge/-Memory-blue)      | Memory driver for the jobs, kv, broadcast                                                                                           |
| ![](https://img.shields.io/badge/-Redis-blue)       | Redis driver for the kv, broadcast                                                                                                  |
| ![](https://img.shields.io/badge/-Boltdb-blue)      | Provides support for the [BoltDB](https://github.com/etcd-io/bbolt) key/value store. Used in the `Jobs` and `KV`                    |

| Plugin                                              | Description                                                                                                                         | Docs                          |
| --------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| ![](https://img.shields.io/badge/-Config-green)     | Provides configuration parsing support to the all plugins                                                                           | [Docs](config/docs/config.md) |
| ![](https://img.shields.io/badge/-GRPC-green)       | Provides GRPC support                                                                                                               | [Docs](grpc/docs/grpc.md)     |
| ![](https://img.shields.io/badge/-Informer-green)   | Provides statistic grabbing capabilities (workers,jobs stat)                                                                        |
| ![](https://img.shields.io/badge/-Broadcast-green)  | Provides broadcasting capabilities to the RR2 via different drivers                                                                 |
| ![](https://img.shields.io/badge/-Logger-green)     | Central logger plugin. Implemented via Uber.zap logger, but supports other loggers.                                                 | [Docs](logger/docs/logger.md) |
| ![](https://img.shields.io/badge/-Metrics-green)    | Provides support for the metrics via [Prometheus](https://prometheus.io/)                                                           |
| ![](https://img.shields.io/badge/-Reload-green)     | Reloads workers on the file changes. Use only for the development                                                                   |
| ![](https://img.shields.io/badge/-Resetter-green)   | Provides support for the `./rr reset` command. Reloads workers pools                                                                |
| ![](https://img.shields.io/badge/-RPC-green)        | Provides support for the RPC across all plugins. Collects `RPC() interface{}` methods and exposes them via RPC                      |
| ![](https://img.shields.io/badge/-Server-green)     | Provides support for the command. Prepare PHP processes                                                                             | [Docs](server/docs/server.md) |
| ![](https://img.shields.io/badge/-Service-green)    | Provides support for the external scripts, binaries which might be started like a service (behaves similar to the systemd services) |
| ![](https://img.shields.io/badge/-Status-green)     | Provides support for the health and readiness checks                                                                                |
| ![](https://img.shields.io/badge/-Websockets-green) | Provides support for the broadcasting events via websockets                                                                         |
| ![](https://img.shields.io/badge/-TCP-green) | Provides support for the raw TCP payloads and TCP connections                                                                         | [Docs](tcp/docs/tcp.md)

Legend:  
`Green` - is a regular RR2 plugins.  
`Blue` - is a driver for the RR2 plugin.

# Writing Plugins

RoadRunner uses Endure container to manage dependencies. This approach is similar to the PHP Container implementation
with automatic method injection. You can create your own plugins, event listeners, middlewares, etc.

To define your plugin, create a struct with public `Init` method with error return value (you can use `spiral/errors` as
the `error` package):

```go
package custom

const PluginName = "custom"

type Plugin struct{}

func (s *Plugin) Init() error {
	return nil
}
```

### Dependencies

You can access other RoadRunner plugins by requesting dependencies in your `Init` method:

```go
package custom

import (
	"github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/rpc"
)

type Service struct{}

func (s *Service) Init(r *rpc.Plugin, rr *http.Plugin) error {
	return nil
}
```

Or collect all plugins implementing particular interface via the `Collects` interface implementation, like this:

```go

package custom

import (
	"github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/rpc"
)

type Middleware interface {
	Middleware(f http.Handler) http.Handler
}

type middleware map[string]Middleware

type Service struct {
	mdwr middleware
}

// Init will be called BEFORE Collects
func (p *Service) Init(r *rpc.Plugin, rr *http.Plugin) error {
	p.mdwr = make(map[string]Middleware)
	return nil
}

// Collects is a special endure interface. Endure analyze all args in the returning methods and searches for the plugins implementing them.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.AddMiddleware,
	}
}

// Here is we are searching for the plugins implementing `endure.Named` AND `Middleware` interfaces.
func (p *Plugin) AddMiddleware(name endure.Named, m Middleware) {
	p.mdwr[name.Name()] = m
}
```

**Make sure to request dependency via pointer.**

### Configuration

In most of the cases, your services would require a set of configuration values. RoadRunner can automatically populate
and validate your configuration structure using `config` plugin (should be imported only config.Configurer interface):

Config sample:

```yaml
custom:
  address: tcp://localhost:8888
```

Plugin:

```go
package custom

import (
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/rpc"

	"github.com/spiral/errors"
)

// Your custom plugin name
const PluginName = "custom"

type Config struct {
	Address string `mapstructure:"address"`
}

type Plugin struct {
	cfg *Config
}

// You can also initialize some defaults values for config keys
func (cfg *Config) InitDefaults() {
	if cfg.Address == "" {
		cfg.Address = "tcp://localhost:8088"
	}
}

func (s *Plugin) Init(r *rpc.Plugin, h *http.Plugin, cfg config.Configurer) error {
	const op = errors.Op("custom_plugin_init") // error operation name
	// Check if the `custom` section exists in the configuration.
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	// Populate the configuration structure
	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		// Error will stop execution
		return errors.E(op, err)
	}

	return nil
}

```

`errors.Disabled` is the special kind of error that tells Endure to disable this plugin and all dependencies of this
root. The RR2 will continue to work after this error type if at least one plugin stay alive.

### Serving

Create `Serve` and `Stop` method in your structure to let RoadRunner start and stop your service.

```go
type Plugin struct {}

func (s *Plugin) Serve() chan error {
const op = errors.Op("custom_plugin_serve")
errCh := make(chan error, 1)

err := s.DoSomeWork()
if err != nil {
// notify endure, that the error occured
errCh <- errors.E(op, err)
return errCh
}
return nil
}

func (s *Plugin) Stop() error {
return s.stopServing()
}

// You may start some listener here
func (s *Plugin) DoSomeWork() error {
return nil
}
```

`Serve` method is thread-safe. It runs in the separate goroutine which managed by the `Endure` container. The one note,
is that you should unblock it when call `Stop` on the container. Otherwise, service will be killed after timeout (can be
set in Endure).

### Collecting dependencies in runtime

RR2 provide a way to collect dependencies in runtime via `Collects` interface. This is very useful for the middlewares
or extending plugins with additional functionality w/o changing it. Let's create an HTTP middleware:

Steps (sample based on the actual `http` plugin and `Middleware` interface):

1. Declare a required interface

```go
// Middleware interface
type Middleware interface {
Middleware(f http.Handler) http.HandlerFunc
}
```

2. Implement method, which should have as an arguments name (`endure.Named` interface) and `Middleware` (step 1).

```go
// Collects collecting http middlewares
func (p *Plugin) AddMiddleware(name endure.Named, m Middleware) {
p.mdwr[name.Name()] = m
}
```

3. Implement `Collects` endure interface for the required structure and return implemented on the step 2 method.

```golang
// Collects collecting http middlewares
func (p *Plugin) Collects() []interface{} {
return []interface{}{
// Endure will analyze the arguments of this function.
p.AddMiddleware,
}
}
```

Endure will automatically check that the registered structure implements all the arguments for the `AddMiddleware`
method (or will find a structure if the argument is structure). In our case, a structure should implement `endure.Named`
interface (which returns user friendly name for the plugin) and `Middleware` interface.

### RPC Methods

You can expose a set of RPC methods for your PHP workers also by using Endure `Collects` interface. Endure will
automatically get the structure and expose RPC method under the `PluginName` name.

```go
func (p *Plugin) Name() string {
return PluginName
}
```

To extend your plugin with RPC methods, the plugin itself will not be changed at all. The only 1 thing to do is to
create a file with RPC methods (let's call it `rpc.go`) and expose here all RPC methods for the plugin:

I assume we created a file `rpc.go`. The next step is to create a structure:

1. Create a structure: (logger is optional)

```go
package custom

type rpc struct {
	srv *Plugin
	log logger.Logger
}
```

2. Create a method, which you want to expose (or multiply methods). BTW, you can use protobuf in the RPC methods:

```go
func (s *rpc) Hello(input string, output *string) error {
*output = input
return nil
}
```

3. Create a method in your plugin (typically `plugin.go`) with the following method signature:

```go
// RPCService returns associated rpc service.
func (p *Plugin) RPC() interface{} {
return &rpc{srv: p, log: p.log}
}
```

4. RPC plugin Collects all plugins which implement `RPC` interface and `endure.Named` automatically (if exists). RPC
   interface accepts no arguments, but returns interface (plugin).

To use it within PHP using `RPC` [instance](/beep-beep/rpc.md):

```php
var_dump($rpc->call('custom.Hello', 'world'));
```

## License:

The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information. Maintained
by [Spiral Scout](https://spiralscout.com).

## Contributors

Thanks to all the people who already contributed!

<a href="https://github.com/spiral/roadrunner/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=spiral/roadrunner" />
</a>
