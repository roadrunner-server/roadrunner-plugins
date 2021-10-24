- ### NewRelic

```yaml
http:
  address: 127.0.0.1:55555
  max_request_size: 1024
  access_logs: false
  middleware: [ "new_relic" ]
  new_relic:
    app_name: "app"
    license_key: "key"

  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s
```

License key and application name could be set via environment variables: (leave `app_name` and `license_key` empty)

- license_key: `NEW_RELIC_LICENSE_KEY`.
- app_name: `NEW_RELIC_APP_NAME`.

To set the New Relic attributes, the PHP worker should send headers values witing the `rr_newrelic` header key.
Attributes should be separated by the `:`, for example `foo:bar`, where `foo` is a key and `bar` is a value. New Relic
attributes sent from the worker will not appear in the HTTP response, they will be sent directly to the New Relic.

To see the sample of the PHP library, see the @arku31 implementation: https://github.com/arku31/roadrunner-newrelic

The special key which PHP may set to overwrite the transaction name is: `transaction_name`. For
example: `transaction_name:foo` means: set transaction name as `foo`. By default, `RequestURI` is used as the
transaction name.

### Custom PHP Response

```php
        $resp = new \Nyholm\Psr7\Response();
        $rrNewRelic = [
            'shopId:1', //custom data
            'auth:password', //custom data
            'transaction_name:test_transaction' //name - special key to override the name. By default it will use requestUri.
        ];

        $resp = $resp->withHeader('rr_newrelic', $rrNewRelic);
```

Where:
- `shopId:1`, `auth:password` - is a custom data that should be attached to the RR New Relic transaction.
- `transaction_name` - is a special header type to overwrite the default transaction name (`RequestURI`).