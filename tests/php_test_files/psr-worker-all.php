<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;
use Spiral\RoadRunner\Jobs\Consumer;
use Spiral\RoadRunner\WorkerInterface;

ini_set('display_errors', 'stderr');
require __DIR__."/vendor/autoload.php";

$env = RoadRunner\Environment::fromGlobals();
$worker = new RoadRunner\Worker(new Goridge\StreamRelay(STDIN, STDOUT));

$worker = match ($env->getMode()) {
    RoadRunner\Environment\Mode::MODE_HTTP => new class($worker) {
        private RoadRunner\Http\PSR7Worker $worker;

        public function __construct(WorkerInterface $worker)
        {
            $this->worker = new RoadRunner\Http\PSR7Worker(
                $worker,
                new \Nyholm\Psr7\Factory\Psr17Factory(),
                new \Nyholm\Psr7\Factory\Psr17Factory(),
                new \Nyholm\Psr7\Factory\Psr17Factory()
            );
        }

        public function run()
        {
            while ($req = $this->worker->waitRequest()) {
                try {
                    $resp = new \Nyholm\Psr7\Response();
                    $resp->getBody()->write("hello world");

                    $this->worker->respond($resp);
                } catch (\Throwable $e) {
                    $this->worker->getWorker()->error((string)$e);
                }
            }
        }
    },

    RoadRunner\Environment\Mode::MODE_JOBS => new class($worker) {
        private Consumer $consumer;

        public function __construct(WorkerInterface $worker)
        {
            $this->consumer = new Consumer();
        }

        public function run()
        {
            while ($task = $this->consumer->waitTask()) {
                $task->complete();
            }
        }
    },

    RoadRunner\Environment\Mode::MODE_TCP => new class($worker) {
        private RoadRunner\Tcp\TcpWorker $worker;

        public function __construct(WorkerInterface $worker)
        {
            $this->worker = new RoadRunner\Tcp\TcpWorker($worker);
        }

        public function run()
        {
            while ($request = $this->worker->waitRequest()) {
                $this->worker->close();
            }
        }
    },

    RoadRunner\Environment\Mode::MODE_GRPC => new class($worker) {
        private RoadRunner\Http\PSR7Worker $worker;

        public function __construct(WorkerInterface $worker)
        {
            $this->worker = new RoadRunner\Http\PSR7Worker(
                $worker,
                new \Nyholm\Psr7\Factory\Psr17Factory(),
                new \Nyholm\Psr7\Factory\Psr17Factory(),
                new \Nyholm\Psr7\Factory\Psr17Factory()
            );
        }

        public function run()
        {
            while ($req = $this->worker->waitRequest()) {
                try {
                    $resp = new \Nyholm\Psr7\Response();
                    $resp->getBody()->write("hello world");

                    $this->worker->respond($resp);
                } catch (\Throwable $e) {
                    $this->worker->getWorker()->error((string)$e);
                }
            }
        }
    },
    default =>  new class($worker) {
        public function __construct(WorkerInterface $worker)
        {
        }

        public function run()
        {
            // do noting
        }
    },
};


$worker->run();