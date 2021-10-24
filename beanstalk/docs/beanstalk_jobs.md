
### Beanstalk Driver

Beanstalk is a simple and fast general purpose work queue. To install Beanstalk,
you can use the [local queue server](https://github.com/beanstalkd/beanstalkd)
or run the server inside [AWS Elastic](https://aws.amazon.com/elasticbeanstalk/).
You can choose any option that is convenient for you.

Setting up the server is similar to setting up AMQP and requires specifying the
connection in the `"beanstalk"` section of your RoadRunner configuration file.

```yaml
beanstalk:
  addr: tcp://127.0.0.1:11300
```

After setting up the connection, you can start using it. Let's take a look at
the complete config with all the options for this driver:

```yaml
beanstalk:
  # Optional section.
  # Default: tcp://127.0.0.1:11300
  addr: tcp://127.0.0.1:11300

  # Optional section.
  # Default: 30s
  timeout: 10s

jobs:
  pipelines:
    # User defined name of the queue.
    example:
      # Required section.
      # Should be "beanstalk" for the Beanstalk driver.
      driver: beanstalk
      
      # Optional section.
      # Default: 10
      priority: 10

      # Optional section.
      # Default: 1
      tube_priority: 1
      
      # Optional section.
      # Default: default
      tube: default

      # Optional section.
      # Default: 5s
      reserve_timeout: 5s
```

These are all settings that are available to you for configuring this type of
driver. Let's take a look at what they are responsible for:
- `priority` - Similar to the same option in other drivers. This is queue
  default priority for for each task pushed into this queue if the priority
  value for these tasks was not explicitly set.

- `tube_priority` - The value for specifying the priority within Beanstalk is
  the internal priority of the server. The value should not exceed `int32` size.

- `tube` - The name of the inner "tube" specific to the Beanstalk driver.