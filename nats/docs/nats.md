### PHP client
- https://github.com/spiral/roadrunner-jobs

### Configuration

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

Driver uses latest JetSteam NATS API.

Legend:
- `subject` - nats [subject](https://docs.nats.io/nats-concepts/subjects). 
- `stream` - nats stream name.
- `deliver_new` - deliver messages since connected to the server.
- `rate_limit` - NATS rate [limiter](https://docs.nats.io/jetstream/concepts/consumers#ratelimit)
- `delete_stream_on_stop` - delete stream on when pipeline stopped.
- `delete_after_ack` - delete message after it successfully acknowledged.

### Worker

```php
<?php

use Spiral\RoadRunner\Jobs\Jobs;

require __DIR__ . '/vendor/autoload.php';

// Jobs service
$jobs = new Jobs(RPC::create('tcp://127.0.0.1:6001'));

// Select "test" queue from jobs
$queue = $jobs->connect('test');

// Create task prototype with default headers
$prototype = $queue->create('echo')
    ->withHeader('attempts', 4)
    ->withHeader('retry-delay', 10)
;

// Execute "echo" task with Closure as payload
$task = $queue->dispatch(
    $prototype->withValue(static fn($arg) => print $arg)
);

var_dump($task->getId() . ' has been queued');

```

