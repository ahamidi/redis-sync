## Redis Sync

Simple tool to replicate one Redis to another.

### How do I use it?

### How does it work?

By default, redis-sync will call `SCAN` on the source Redis and iteratively 
`DUMP` and `RESTORE` those keys in the target Redis in batches.

### Limitations

* Since `SCAN` is used an active Redis can result in an inconsistent replica.
    It is recommended that writes (and probably reads) should be disabled before
    attempting to sync.
* Since `DUMP` is used, the serialized value does not contain expire information.
    You can read more about this limitation in the [Redis Docs](https://redis.io/commands/dump)

