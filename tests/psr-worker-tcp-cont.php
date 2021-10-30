<?php

require __DIR__ . '/vendor/autoload.php';

use Spiral\RoadRunner\Worker;
use Spiral\RoadRunner\Tcp\TcpWorker;

// Create new RoadRunner worker from global environment
$worker = Worker::create();

$tcpWorker = new TcpWorker($worker);

while ($request = $tcpWorker->waitRequest()) {

    try {
        if ($request->event === TcpWorker::EVENT_CONNECTED) {
            // -----------------

            // Or send response to the TCP connection, for example, to the SMTP client
            $tcpWorker->respond("hello \r\n");

        } else if ($request->event === TcpWorker::EVENT_DATA) {

            $body = $request->body;
                // Send response to the TCP connection and wait for the next request
                $tcpWorker->respond(json_encode([
                    'remote_addr' => $request->remoteAddr,
                    'server' => $request->server,
                    'uuid' => $request->connectionUuid,
                    'body' => $request->body,
                    'event' => $request->event
                ]));

        // Handle closed connection event
        } else if ($request->event === TcpWorker::EVENT_CLOSED) {
        }

    } catch (\Throwable $e) {
        $tcpWorker->respond("Something went wrong\r\n", true);
        $worker->error((string)$e);
    }
}
