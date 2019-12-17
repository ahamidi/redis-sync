## Redis Sync

Simple tool to replicate one Redis to another.

### Download Binary or Build

You can find binaries for all major platforms under [Releases](https://github.com/ahamidi/redis-sync/releases)

Alternatively you can build from source using the following command:
```
go get -u . && go build ./...
```

### How do I use it?
```
Usage of redis-sync:
  -read-concurrency int
    	source Redis read concurrency (default 10)
  -replace
    	replace existing keys on target
  -source string
    	source Redis host (default localhost:6379) (default "redis://127.0.0.1:6379")
  -source-database int
    	source Redis DB (default 0)
  -target string
    	target Redis database
  -target-database int
    	target Redis DB (default 0)
  -write-concurrency int
    	target Redis write concurrency (default 10)
  -sync-ttl
    	copy TTL times (default false)
```

**Example**

```
# Sync local redis db 0 to remote Redis db 0
redis-sync --target redis://username:password@192.168.1.100:6397
```

### How does it work?

By default, redis-sync will call `SCAN` on the source Redis and iteratively 
`DUMP` and `RESTORE` those keys in the target Redis in batches.

### Limitations

* Since `SCAN` is used an active Redis can result in an inconsistent replica.
    It is recommended that writes (and probably reads) should be disabled before
    attempting to sync.
* Since `DUMP` is used, the serialized value does not contain expire information.
    You can read more about this limitation in the [Redis Docs](https://redis.io/commands/dump)

