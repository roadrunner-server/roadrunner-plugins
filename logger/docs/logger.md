## PHP Client

[Roadrunner Worker](https://github.com/spiral/roadrunner-worker)

## Configuration

```yaml
# Logging settings (docs: https://roadrunner.dev/docs/beep-beep-logging)
logs:
  # Logging mode can be "development", "production" or "raw". Do not forget to change this value for production environment.
  #
  # Development mode (which makes DPanicLevel logs panic), uses a console encoder, writes to standard error, and
  # disables sampling. Stacktraces are automatically included on logs of WarnLevel and above.
  #
  # Default: "development"
  mode: development

  # Logging level can be "panic", "error", "warn", "info", "debug".
  #
  # Default: "debug"
  level: debug

  # Encoding format can be "console" or "json" (last is preferred for production usage).
  #
  # Default: "console"
  encoding: console

  # Output can be file (eg.: "/var/log/rr_errors.log"), "stderr" or "stdout".
  #
  # Default: "stderr"
  output: stderr

  # Errors only output can be file (eg.: "/var/log/rr_errors.log"), "stderr" or "stdout".
  #
  # Default: "stderr"
  err_output: stderr

  # You can configure each plugin log messages individually (key is plugin name, and value is logging options in same
  # format as above).
  #
  # Default: <empty map>
  channels:
    http:
      mode: development
      level: panic
      encoding: console
      output: stdout
      err_output: stderr
    server:
      mode: production
      level: info
      encoding: json
      output: stdout
      err_output: stdout
    rpc:
      mode: raw
      level: debug
      encoding: console
      output: stderr
      err_output: stdout
```

## Minimal dependencies:

1. `Config` plugin to read and populate plugin's configuration.

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

1. Use development logger mode only for the dev. It uses more resources even with `panic` level and not suitable for the production.
2. You can configure log parameters for the every plugin you use.

## Common issues:
