<?php

/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;
use Spiral\Goridge\StreamRelay;

require __DIR__ . "/vendor/autoload.php";

$rr = new RoadRunner\Worker(new StreamRelay(\STDIN, \STDOUT));

while ($in = $rr->waitPayload()) {
    try {
        $rr->respond(new RoadRunner\Payload('foo'));
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
