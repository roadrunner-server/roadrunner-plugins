Response protocol used to communicate between worker and RR. When a worker completes its job, it should send a typed
response. The response should contain:

1. `type` field with the message type. Can be treated as enums.
2. `data` field with the dynamic response related to the type.

Types are:

```
0 - NO_ERROR
1 - ERROR
2 - RESPONSE
```

#### `NO_ERROR`: contains only `type` - 0, and empty `data`.  
example:
```json
{
  "type": 0,
  "data": {}
}
```
---

#### `ERROR` : contains `type` - 1, and `data` field with: 
- `message` describing the error.  
- `requeue` flag to requeue the job.  
- `delay_seconds`: to delay a queue for a provided amount of seconds.   
- `headers` - job's headers represented as hashmap with string key and array of strings as a value.  

example:
```json
{
  "type": 1,
  "data": {
    "message": "internal worker error",
    "requeue": true,
    "headers": [
      {
        "test": [
          "1",
          "2",
          "3"
        ]
      }
    ],
    "delay_seconds": 10
  }
}
```


---

#### `RESPONSE`: response message type. Message type - 2, with:

- `queue` (eg. tube, subject) name.
- `payload` which should be sent via response `Context`.   

example:
```json
{
  "type": 2,
  "data": {
    "queue": "foo",
    "payload": "binary_payload"
  }
}
```
