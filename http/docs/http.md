### Configuration

```json
{
  "http": {
    "address": "127.0.0.1:8080",
    "max_request_size": 256,
    "middleware": [
      "headers",
      "gzip"
    ],
    "ssl": {
      "address": "0.0.0.0:443",
      "acme": {
        "certs_dir": "rr_le_certs",
        "email": "you-email-here@email",
        "challenge_type": "http-01",
        "use_production_endpoint": true,
        "domains": [
          "your-cool-domains.here"
        ]
      }
    }
  },
  "fcgi": {
    "address": "tcp://0.0.0.0:7921"
  },
  "http2": {
    "h2c": false,
    "max_concurrent_streams": 128
  },
  "trusted_subnets": [
    "10.0.0.0/8",
    "127.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
    "::1/128",
    "fc00::/7",
    "fe80::/10"
  ],
  "uploads": {
    "dir": "/tmp",
    "forbid": [
      ".php",
      ".exe",
      ".bat",
      ".sh"
    ]
  },
  "headers": {
    "cors": {
      "allowed_origin": "*",
      "allowed_headers": "*",
      "allowed_methods": "GET,POST,PUT,DELETE",
      "allow_credentials": true,
      "exposed_headers": "Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma",
      "max_age": 600
    },
    "request": {
      "input": "custom-header"
    },
    "response": {
      "X-Powered-By": "RoadRunner"
    }
  },
  "static": {
    "dir": ".",
    "forbid": [
      ".go"
    ],
    "allow": [
      ".txt",
      ".php"
    ],
    "calculate_etag": false,
    "weak": false,
    "request": {
      "input": "custom-header"
    },
    "response": {
      "output": "output-header"
    }
  },
  "pool": {
    "debug": false,
    "num_workers": 0,
    "max_jobs": 64,
    "allocate_timeout": "60s",
    "destroy_timeout": "60s",
    "supervisor": {
      "watch_tick": "1s",
      "ttl": "0s",
      "idle_ttl": "10s",
      "max_worker_memory": 128,
      "exec_ttl": "60s"
    }
  }
}
}
```

- YAML configuration is [here](https://github.com/spiral/roadrunner-binary/blob/master/.rr.yaml#L373)

#### Configuration tips:

- If you use ACME provider to obtain certificates, you only need to specify SSL address in the root configuration.
- There is no certificates auto-renewal support yet, but this feature planned for the future. To renew you certificates,
  just re-run RR with `obtain_certificates` set to
  true ([link](https://letsencrypt.org/docs/faq/#what-is-the-lifetime-for-let-s-encrypt-certificates-for-how-long-are-they-valid))
  .

### Worker

```php
<?php

require __DIR__ . '/vendor/autoload.php';

use Nyholm\Psr7\Response;
use Nyholm\Psr7\Factory\Psr17Factory;

use Spiral\RoadRunner\Worker;
use Spiral\RoadRunner\Http\PSR7Worker;


// Create new RoadRunner worker from global environment
$worker = Worker::create();

// Create common PSR-17 HTTP factory
$factory = new Psr17Factory();

//
// Create PSR-7 worker and pass:
//  - RoadRunner worker
//  - PSR-17 ServerRequestFactory
//  - PSR-17 StreamFactory
//  - PSR-17 UploadFilesFactory
//
$psr7 = new PSR7Worker($worker, $factory, $factory, $factory);

while (true) {
    try {
        $request = $psr7->waitRequest();
    } catch (\Throwable $e) {
        // Although the PSR-17 specification clearly states that there can be
	// no exceptions when creating a request, however, some implementations
	// may violate this rule. Therefore, it is recommended to process the 
	// incoming request for errors.
        //
        // Send "Bad Request" response.
        $psr7->respond(new Response(400));
        continue;
    }

    try {
        // Here is where the call to your application code will be located. 
	// For example:
	//
        //  $response = $app->send($request);
        //
        // Reply by the 200 OK response
        $psr7->respond(new Response(200, [], 'Hello RoadRunner!'));
    } catch (\Throwable $e) {
        // In case of any exceptions in the application code, you should handle
	// them and inform the client about the presence of a server error.
	//
        // Reply by the 500 Internal Server Error response
        $psr7->respond(new Response(500, [], 'Something Went Wrong!'));
        
	// Additionally, we can inform the RoadRunner that the processing 
	// of the request failed.
        $worker->error((string)$e);
    }
}

```

## Useful links

- [Error handling](https://github.com/spiral/roadrunner-docs/blob/master/php/error-handling.md)
- [HTTPS](https://github.com/spiral/roadrunner-docs/blob/master/http/https.md)
- [Static content](https://github.com/spiral/roadrunner-docs/blob/master/http/static.md)
- [Golang http middleware](https://github.com/spiral/roadrunner-docs/blob/master/http/middleware.md)