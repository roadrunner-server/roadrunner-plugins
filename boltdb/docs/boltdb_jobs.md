### Local (based on the boltdb) Driver

This type of driver is already supported by the RoadRunner and does not require
any additional installations. It uses boltdb as its main storage for the jobs. This driver should be used locally, for
testing or developing purposes. It can be used in the production, but this type of driver can't handle
huge load. Maximum RPS it can have no more than 30-50.

Data in this driver persists in the boltdb database file. You can't open same file simultaneously for the 2 pipelines or
for the KV plugin and Jobs plugin. This is boltdb limitation on concurrent access from the 2 processes to the same file.

The complete `boltdb` driver configuration looks like this:

```yaml

boltdb:
  permissions: 0777

jobs:
  pipelines:
    # User defined name of the queue.
    example:
      # Required section.
      # Should be "boltdb" for the local driver.
      driver: boltdb
      
      # Optional section.
      # Default: 10
      priority: 10
      
      # Optional section.
      # Default: 1000
      prefetch: 1000
```

Below is a more detailed description of each of the in-memory-specific options:
- `priority` - Queue default priority for each task pushed into this queue
  if the priority value for these tasks was not explicitly set.

- `prefetch` - A local buffer between the PQ (priority queue) and driver. If the
  PQ size is set to 100 and prefetch to 100000, you'll be able to push up to
  prefetch number of jobs even if PQ is full.

- `file` - boltdb database file to use. Might be a full path with file: `/foo/bar/rr1.db`. Default: `rr.db`. 