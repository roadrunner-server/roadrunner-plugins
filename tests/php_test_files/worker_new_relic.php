<?php
/**
 * @var Goridge\RelayInterface $relay
 */
use Spiral\Goridge;
use Spiral\RoadRunner;
use Arku\Newrelic\Response\EnrichResponse;
use Arku\Newrelic\Transactions\Segment;
use Arku\Newrelic\Transactions\TransactionDetail;
use Arku\Newrelic\Transformers\TransactionDetailTransformer;

ini_set('display_errors', 'stderr');
require __DIR__ . "/vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psr7 = new RoadRunner\Http\PSR7Worker(
    $worker,
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory()
);

while ($req = $psr7->waitRequest()) {
    try {
        $resp = new \Nyholm\Psr7\Response(200, [], 'Hello RoadRunner!');
        $resp = $resp->withHeader('Content-type', 'application/json');

        $transactionDetail = new TransactionDetail();
        $transactionDetail->setName('test');
        $transactionDetail->setCustomData('key', 'value');

        $segment = new Segment();
        $segment->setName('testSegment');
        $segment->setDuration('1');
        $segment->setMeta(['testmetakey' => 'testmetavalue']);

        $transactionDetail->addSegment($segment);

        $enricher = new EnrichResponse(new TransactionDetailTransformer());
        $response = $enricher->enrich($resp, $transactionDetail);

        $psr7->respond($response);
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}