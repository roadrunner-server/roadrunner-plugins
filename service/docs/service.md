### Service plugin
- `Since RR v2.0.5`

### Configuration

```yaml
# Service plugin settings
service:
  # User defined service name
  #
  # Default: none, required
  some_service_1:
    # Command to execute. Can be any command here which can be executed.
    #
    # Default: none, required.
    command: php tests/plugins/service/test_files/loop.php

    # Endings for the stderr/stdout output
    #
    # Default: "\n". Available options: any.
    line_ending: "\n"

    # Number of copies (processes) to start per command.
    #
    # Default: 1

    process_num: 1
    # Allowed execute timeout.
    #
    # Default: 0 (infinity), can be 1s, 2m, 2h (seconds, minutes, hours)

    exec_timeout: 0
    # Remain process after exit. In other words, restart process after exit with any exit code.
    #
    # Default: "false"

    remain_after_exit: true
    # Number of seconds to wait before process restart.
    #
    # Default: 30
    restart_sec: 1
    
    # Environment variables to pass to the process
    #
    # Default: empty
    env:
       - "FOO": "BAR"
       - "FOO2": "BAZ"

  # User defined service name
  #
  # Default: none, required
  some_service_2:
    # Command to execute. Can be any command here which can be executed.
    #
    # Default: none, required.
    command: "./some_executable"

    # Endings for the stderr/stdout output
    #
    # Default: "\n". Available options: any.
   
    line_ending: "\n"

    # Number of copies (processes) to start per command.
    #
    # Default: 1
    process_num: 1

    # Allowed execute timeout.
    #
    # Default: 0 (infinity), can be 1s, 2m, 2h (seconds, minutes, hours)
    exec_timeout: 0

    # Remain process after exit. In other words, restart process after exit with any exit code.
    #
    # Default: "false"
    remain_after_exit: true

    # Number of seconds to wait before process restart.
    #
    # Default: 30
    restart_sec: 1
```

### Worker

1. This plugin not limiting the user, how to write a worker. It can be a PHP script, or some executable or binary written in any language.
For example, command could be like `go run main.go` to run some Go file.
Sample of the worker:

```php
<?php
for ($x = 0; $x <= 1000; $x++) {
  sleep(1);
  error_log("The number is: $x", 0);
}
?>
```

### Useful links
- [Service](https://roadrunner.dev/docs/beep-beep-service)
