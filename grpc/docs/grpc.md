## PHP Client

[Roadrunner GRPC](https://github.com/spiral/roadrunner-grpc)

### Compiling proto sample:
```
protoc --plugin=protoc-gen-php-grpc --php-grpc_out=<OUTPUT DIRECTORY> simple.proto
```

- `protoc-gen-php-grpc` plugin can be downloaded from the `roadrunner-binary` releases page.

## Configuration

```yaml
grpc:
  # GRPC address to listen
  #
  # This option is required
  listen: "tcp://localhost:9001"

  # Proto files to use
  #
  # This option is required. At least one proto file must be specified.
  proto:
      - "proto/test/test.proto"
      - "proto/health/health.proto"

  # GRPC TLS configuration
  #
  # This section is optional
  tls:
    # Path to the key file
    #
    # This option is required
    key: ""

    # Path to the certificate
    #
    # This option is required
    cert: ""

    # Path to the CA certificate
    #
    # This option is optional
    root_ca: ""

    # Client auth type
    #
    # This option is optional. Default value: no_client_certs. Possible values: request_client_cert, require_any_client_cert, verify_client_cert_if_given, require_and_verify_client_cert, no_client_certs
    client_auth_type: ""

  # Maximum send message size
  #
  # This option is optional. Default value: 50 (MB)
  max_send_msg_size: 50

  # Maximum receive message size
  #
  # This option is optional. Default value: 50 (MB)
  max_recv_msg_size: 50

  # MaxConnectionIdle is a duration for the amount of time after which an
  #	idle connection would be closed by sending a GoAway. Idleness duration is
  #	defined since the most recent time the number of outstanding RPCs became
  #	zero or the connection establishment.
  #
  # This option is optional. Default value: infinity.
  max_connection_idle: 0s

  # MaxConnectionAge is a duration for the maximum amount of time a
  #	connection may exist before it will be closed by sending a GoAway. A
  #	random jitter of +/-10% will be added to MaxConnectionAge to spread out
  #	connection storms.
  #
  # This option is optional. Default value: infinity.
  max_connection_age: 0s

  # MaxConnectionAgeGrace is an additive period after MaxConnectionAge after
  #	which the connection will be forcibly closed.
  max_connection_age_grace: 0s

  # MaxConnectionAgeGrace is an additive period after MaxConnectionAge after
  #	which the connection will be forcibly closed.
  #
  # This option is optional: Default value: 10
  max_concurrent_streams: 10

  # After a duration of this time if the server doesn't see any activity it
  #	pings the client to see if the transport is still alive.
  #	If set below 1s, a minimum value of 1s will be used instead.
  #
  # This option is optional. Default value: 2h
  ping_time: 1s

  # After having pinged for keepalive check, the server waits for a duration
  #	of Timeout and if no activity is seen even after that the connection is
  #	closed.
  #
  # This option is optional. Default value: 20s
  timeout: 200s

  # Usual workers pool configuration
  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60
```

## Minimal dependencies:

1. `Server` plugin for the workers pool.
2. `Logger` plugin to show log messages.
3. `Config` plugin to read and populate plugin's configuration.

## GRPC worker sample:

```php
<?php

/**
 * Sample GRPC PHP server.
 */

use Service\EchoInterface;
use Spiral\RoadRunner\GRPC\Server;
use Spiral\RoadRunner\Worker;

require __DIR__ . '/vendor/autoload.php';

$server = new Server(null, [
    'debug' => false, // optional (default: false)
]);

$server->registerService(EchoInterface::class, new EchoService());

$server->serve(Worker::create());

```

#### Proto file sample:

```protobuf
syntax = "proto3";

package service;
option go_package = "./;test";

service Test {
    rpc Echo (Message) returns (Message) {
    }

    rpc Throw (Message) returns (Message) {
    }

    rpc Die (Message) returns (Message) {
    }

    rpc Info (Message) returns (Message) {
    }

    rpc Ping (EmptyMessage) returns (EmptyMessage) {
    }
}

message Message {
    string msg = 1;
}

message EmptyMessage {
}

message DetailsMessageForException {
    uint64 code = 1;
    string message = 2;
}
```

Test certificates (including `root ca`) located [here](../../tests/plugins/grpc/configs/test-certs).

## Common issues:
1. Registering two services with the same name is not allowed. GRPC server will panic after that.