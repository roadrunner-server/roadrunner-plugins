<?php

use Spiral\RoadRunner\Jobs\Jobs;
use Spiral\RoadRunner\Jobs\Queue\MemoryCreateInfo;
use Spiral\RoadRunner\Jobs\Consumer;
use Spiral\RoadRunner\Jobs\Task\ReceivedTaskInterface;

require __DIR__ . "/vendor/autoload.php";

$jobs = new Spiral\RoadRunner\Jobs\Jobs(
    // Expects RPC connection
    Spiral\Goridge\RPC\RPC::create('tcp://127.0.0.1:6001')
);

//
// Create a new "example" in-memory queue
//
$queue = $jobs->create(new MemoryCreateInfo(
    name: 'example',
    priority: 42,
    prefetch: 10,
));

$queue->resume();

$consumer = new Consumer();
while ($task = $consumer->waitTask()) {
    $task->complete();
}