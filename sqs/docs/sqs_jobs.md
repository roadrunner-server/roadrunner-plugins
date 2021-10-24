### SQS Driver

[Amazon SQS (Simple Queue Service)](https://aws.amazon.com/sqs/) is an
alternative queue server also developed by Amazon and is also part of the AWS
service infrastructure. If you prefer to use the "cloud" option, then you can
use the [ready-made documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-configuring.html)
for its installation.

In addition to the possibility of using this queue server within the AWS, you
can also use the local installation of this system on your own servers. If you
prefer this option, then you can use [softwaremill's implementation](https://github.com/softwaremill/elasticmq)
of the Amazon SQS server.

After you have created the SQS server, you need to specify the following
connection settings in `sqs` configuration settings. Unlike AMQP and Beanstalk,
SQS requires more values to set up a connection and will be different from what
we're used to:

```yaml
sqs:
  # Required AccessKey ID.
  # Default: empty
  key: access-key

  # Required secret access key.
  # Default: empty
  secret: api-secret

  # Required AWS region.
  # Default: empty
  region: us-west-1

  # Required AWS session token.
  # Default: empty
  session_token: test

  # Required AWS SQS endpoint to connect.
  # Default: http://127.0.0.1:9324
  endpoint: http://127.0.0.1:9324
```

> Please note that although each of the sections contains default values, it is
> marked as "required". This means that in almost all cases they are required to
> be specified in order to correctly configure the driver.

After you have configured the connection - you should configure the queue that
will use this connection:

```yaml
sqs:
  # SQS connection configuration...

jobs:
  pipelines:
    # Required section.
    # Should be "sqs" for the Amazon SQS driver.
    driver: sqs
    
    # Optional section.
    # Default: 10
    prefetch: 10
    
    # Optional section.
    # Default: 0
    visibility_timeout: 0
    
    # Optional section.
    # Default: 0
    wait_time_seconds: 0
    
    # Optional section.
    # Default: default
    queue: default
    
    # Optional section.
    # Default: empty
    attributes:
      DelaySeconds: 42
      # etc... see https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SetQueueAttributes.html
      
    # Optional section.
    # Default: empty
    tags:
      test: "tag"
```

Below is a more detailed description of each of the SQS-specific options:
- `prefetch` - Number of jobs to prefetch from the SQS. Amazon SQS never returns
  more messages than this value (however, fewer messages might be returned).
  Valid values: 1 to 10. Any number bigger than 10 will be rounded to 10.
  Default: `10`.

- `visibility_timeout` - The duration (in seconds) that the received messages
  are hidden from subsequent retrieve requests after being retrieved by a
  ReceiveMessage request. Max value is 43200 seconds (12 hours). Default: `0`.

- `wait_time_seconds` - The duration (in seconds) for which the call waits for
  a message to arrive in the queue before returning. If a message is available,
  the call returns sooner than WaitTimeSeconds. If no messages are available and
  the wait time expires, the call returns successfully with an empty list of
  messages. Default: `5`.

- `queue` - SQS internal queue name. Can contain alphanumeric characters,
  hyphens (-), and underscores (_). Default value is `"default"` string.

- `attributes` - List of the [AWS SQS attributes](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SetQueueAttributes.html).
> For example
> ```yaml
> attributes:
>   DelaySeconds: 0
>   MaximumMessageSize: 262144
>   MessageRetentionPeriod: 345600
>   ReceiveMessageWaitTimeSeconds: 0
>   VisibilityTimeout: 30
> ```

- `tags` - Tags don't have any semantic meaning. Amazon SQS interprets tags as
  character.
> Please note that this functionality is rarely used and slows down the work of
> queues: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-queue-tags.html