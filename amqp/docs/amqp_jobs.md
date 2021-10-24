### AMQP Driver

Strictly speaking, AMQP (and 0.9.1 version used) is a protocol, not a full-fledged driver, so you can use
any servers that support this protocol (on your own, only rabbitmq was tested) , such as:
[RabbitMQ](https://www.rabbitmq.com/), [Apache Qpid](http://qpid.apache.org/) or
[Apache ActiveMQ](http://activemq.apache.org/). However, it is recommended to
use RabbitMQ as the main implementation, and reliable performance with other
implementations is not guaranteed.

To install and configure the RabbitMQ, use the corresponding
[documentation page](https://www.rabbitmq.com/download.html). After that, you
should configure the connection to the server in the "`amqp`" section. This
configuration section contains exactly one `addr` key with a
[connection DSN](https://www.rabbitmq.com/uri-spec.html).

```yaml
amqp:
  addr: amqp://guest:guest@localhost:5672
```

After creating a connection to the server, you can create a new queue that will
use this connection and which will contain the queue settings (including
amqp-specific):

```yaml
amqp:
  addr: amqp://guest:guest@localhost:5672


jobs:
  pipelines:
    # User defined name of the queue.
    example:
      # Required section.
      # Should be "amqp" for the AMQP driver.
      driver: amqp
      
      # Optional section.
      # Default: 10
      priority: 10
      
      # Optional section.
      # Default: 100
      prefetch: 100

      # Optional section.
      # Default: "default"
      queue: "default"

      # Optional section.
      # Default: "amqp.default"
      exchange: "amqp.default"

      # Optional section.
      # Default: "direct"
      exchange_type: "direct"

      # Optional section.
      # Default: "" (empty)
      routing_key: ""

      # Optional section.
      # Default: false
      exclusive: false

      # Optional section.
      # Default: false
      multiple_ack: false

      # Optional section.
      # Default: false
      requeue_on_fail: false
```

Below is a more detailed description of each of the amqp-specific options:
- `priority` - Queue default priority for for each task pushed into this queue
  if the priority value for these tasks was not explicitly set.

- `prefetch` - The client can request that messages be sent in advance so that
  when the client finishes processing a message, the following message is
  already held locally, rather than needing to be sent down the channel.
  Prefetching gives a performance improvement. This field specifies the prefetch
  window size in octets. See also ["prefetch-size"](https://www.rabbitmq.com/amqp-0-9-1-reference.html)
  in AMQP QoS documentation reference.

- `queue` - AMQP internal (inside the driver) queue name.

- `exchange` - The name of AMQP exchange to which tasks are sent. Exchange
  distributes the tasks to one or more queues. It routes tasks to the queue
  based on the created bindings between it and the queue. See also
  ["AMQP model"](https://www.rabbitmq.com/tutorials/amqp-concepts.html#amqp-model)
  documentation section.

- `exchange_type` - The type of task delivery. May be one of `direct`, `topics`,
  `headers` or `fanout`.
    - `direct` - Used when a task needs to be delivered to specific queues. The
      task is published to an exchanger with a specific routing key and goes to
      all queues that are associated with this exchanger with a similar routing
      key.
    - `topics` - Similarly, `direct` exchange enables selective routing by
      comparing the routing key. But, in this case, the key is set using a
      template, like: `user.*.messages`.
    - `fanout` - All tasks are delivered to all queues even if a routing key is
      specified in the task.
    - `headers` - Routes tasks to related queues based on a comparison of the
      (key, value) pairs of the headers property of the binding and the similar
      property of the message.

    - `routing_key` - Queue's routing key.

    - `exclusive` - Exclusive queues can't be redeclared. If set to true and
      you'll try to declare the same pipeline twice, that will lead to an error.

    - `multiple_ack` - This delivery and all prior unacknowledged deliveries on
      the same channel will be acknowledged. This is useful for batch processing
      of deliveries. Applicable only for the Ack, not for the Nack.

    - `requeue_on_fail` - Requeue on Nack.