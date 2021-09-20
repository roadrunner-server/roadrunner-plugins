## PHP Client

[Roadrunner Worker](https://github.com/spiral/roadrunner-worker)

## Configuration

This plugin parse other's plugins configuration and uses flags to find the YAML config.

## Flags

1. `-c [PATH]`: points RR to the configuration location.

## Minimal dependencies:

This plugin is independent.

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

1. By default, `.rr.yaml` used as the configuration, located in the same directory with RR binary.

## Common issues:
