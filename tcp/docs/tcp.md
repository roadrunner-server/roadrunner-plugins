# TCP Plugin

### Client library
- https://github.com/spiral/roadrunner-tcp

### Configuration

```yaml
tcp:
  servers:
    tcp_access_point_1:
      addr: tcp://127.0.0.1:7777
      delimiter: "\r\n"
    server2:
      addr: tcp://127.0.0.1:8889
      read_buf_size: 10
    server3:
      addr: tcp://127.0.0.1:8810
      delimiter: "\r\n"
      read_buf_size: 1
```

### Protocol

### Worker sample

```php
<?php

require __DIR__ . '/vendor/autoload.php';

use Spiral\RoadRunner\Worker;
use Spiral\RoadRunner\Tcp\TcpWorker;

// Create new RoadRunner worker from global environment
$worker = Worker::create();

$tcpWorker = new TcpWorker($worker);

while ($request = $tcpWorker->waitRequest()) {

    try {
        if ($request->event === TcpWorker::EVENT_CONNECTED) 
            // You can close connection according your restrictions
            if ($request->remoteAddr !== '127.0.0.1') {
                $tcpWorker->close();
                continue;
            }
            
            // -----------------
            
            // Or continue read data from server
            // By default, server closes connection if a worker doesn't send CONTINUE response 
            $tcpWorker->read();
            
            // -----------------
            
            // Or send response to the TCP connection, for example, to the SMTP client
            $tcpWorker->respond("220 mailamie \r\n");
            
        } elseif ($request->event === TcpWorker::EVENT_DATA) {
                   
            $body = $request->body;
            
            // ... handle request from TCP server [tcp_access_point_1]
            if ($request->server === 'tcp_access_point_1') {

                // Send response and close connection
                $tcpWorker->respond('Access denied', true);
               
            // ... handle request from TCP server [server2] 
            } elseif ($request->server === 'server2') {
                
                // Send response to the TCP connection and wait for the next request
                $tcpWorker->respond(json_encode([
                    'remote_addr' => $request->remoteAddr,
                    'server' => $request->server,
                    'uuid' => $request->connectionUuid,
                    'body' => $request->body,
                    'event' => $request->event
                ]));
            }
           
        // Handle closed connection event 
        } elseif ($request->event === TcpWorker::EVENT_CLOSED) {
            // Do something ...
            
            // You don't need to send response on closed connection
        }
        
    } catch (\Throwable $e) {
        $tcpWorker->respond("Something went wrong\r\n", true);
        $worker->error((string)$e);
    }
}
```