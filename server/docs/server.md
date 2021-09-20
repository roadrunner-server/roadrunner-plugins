## PHP Client

[Roadrunner Worker](https://github.com/spiral/roadrunner-worker)

## Configuration

```yaml
# Application server settings (docs: https://roadrunner.dev/docs/php-worker)
server:
  # Worker starting command, with any required arguments.
  #
  # This option is required.
  command: "php psr-worker.php"

  # Username (not UID) for the worker processes. An empty value means to use the RR process user.
  #
  # Default: ""
  user: ""

  # Group name (not GID) for the worker processes. An empty value means to use the RR process user.
  #
  # Default: ""
  group: ""

  # Environment variables for the worker processes.
  #
  # Default: <empty map>
  env:
    - SOME_KEY: "SOME_VALUE"
    - SOME_KEY2: "SOME_VALUE2"

  # Worker relay can be: "pipes", TCP (eg.: tcp://127.0.0.1:6002), or socket (eg.: unix:///var/run/rr.sock).
  #
  # Default: "pipes"
  relay: pipes

  # Timeout for relay connection establishing (only for socket and TCP port relay).
  #
  # Default: 60s
  relay_timeout: 60s
```

## Minimal dependencies:

1. `Logger` plugin to show log messages.
2. `Configurer` plugin to read and populate plugin's configuration.

## Worker sample:

```php
<?php

require __DIR__ . '/vendor/autoload.php';

// Create a new Worker from global environment
$worker = \Spiral\RoadRunner\Worker::create();

while ($data = $worker->waitPayload()) {
    // Received Payload
    var_dump($data);

    // Respond Answer
    $worker->respond(new \Spiral\RoadRunner\Payload('DONE'));
}
```

## Tips:

1. If your application uses mostly IO (disk, network, etc), you can allocate as much worker as you have memory for the application. Workers are cheap. Hello-world worker consumes no more than ~26Mb of RSS memory.
2. For the CPU bound operation, see a CPU avg. load and choose the number of workers to consume 90-95% CPU. Leave few percents for the GO's GC (not necessary btw).
3. If you have ~const workers latency, you may calculate the number of workers needed to handle target [load](https://github.com/spiral/roadrunner/discussions/799#discussioncomment-1332646).

## Common issues:
