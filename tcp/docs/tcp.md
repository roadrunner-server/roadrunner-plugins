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

Where:
- `servers`: is the list of TCP servers to start. Every server should contain:
  1. `addr`: server address with port.
  2. `delimiter`: data packets delimiter. Every send should end either with EOF or with the delimiter.
  3. `read_buf_size`: chunks that RR uses to read the data. In MB. If you expect big payloads on a TCP server, to reduce `read` syscalls, would be a good practice to use a fairly big enough buffer.

### RR Protocol

Protocol used to have bi-directional communication channel between the PHP worker and the RR server. The protocol can be used by any third-party library and has its own client API with the RR. Our reference implementation is: https://github.com/spiral/roadrunner-tcp.

- `OnConnect`: when the connection is established, RR sends payload with the `CONNECTED` header to the worker with connection uuid, server name, and connection remote address. The PHP worker might then respond with the following headers:
  1. `WRITE` - to write the data in the connection and then start the read loop. After data arrived in the connection, RR will read it and send to it the PHP worker with the header `DATA`. 
  2. `CONTINUE` - to start the read loop without writing the data into the connection. RR will send read data to the PHP worker with the `DATA` header.
  3. `WRITECLOSE` - to write the data, close the connection and free up the allocated resources. RR will send the `CLOSED` header to the worker after the actual data will be written and the connection is closed.
  4. `CLOSE` - to just close the connection and free up the allocated resources. RR will send the `CLOSED` header to the PHP worker.

- `OnConnectionClosed`: when the connection is closed for any reason, RR sends the `CLOSED` header to the worker with the connection UUID, server name and connection remote address.
- `OnDataArrived`: when the data arrived, RR read the data expecting delimiter at the end of the read and sends the data to the PHP worker with the `DATA` header.

To summarize:   
PHP worker sends the following headers:
- `WRITE`
- `CONTINUE`
- `WRITECLOSE`
- `CLOSE`

RR sends:
- `CONNECTED`
- `DATA`
- `CLOSED`


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