package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/garyburd/redigo/redis"
)

const keyReaderBufferSize = 500

type redisKey struct {
	Key   string
	Value string
}

var sourceRedis = flag.String("source", "redis://localhost:6379", "source Redis host (default localhost:6379)")
var sourceDB = flag.Int("source-database", 0, "source Redis DB (default 0)")
var targetRedis = flag.String("target", "", "target Redis database")
var targetDB = flag.Int("target-database", 0, "target Redis DB (default 0)")
var targetPoolSize = flag.Int("write-concurrency", 10, "target Redis write concurrency")
var replaceKeys = flag.Bool("replace", false, "replace existing keys on target")

func main() {
	flag.Parse()

	// Connect to source
	srcRedis, err := redis.DialURL(*sourceRedis)
	if err != nil {
		log.Fatal(err)
	}

	// Select database
	if *sourceDB != 0 {
		srcRedis.Do("SELECT", *sourceDB)
	}

	// Read
	keyChan, _ := keyReader(srcRedis)

	// Connect to target
	targetPool := newPoolFromURL(*targetRedis, *targetPoolSize)

	// Setup spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Syncing Redis Databases "
	s.Start()

	wg := &sync.WaitGroup{}
	wg.Add(*targetPoolSize)
	for i := 0; i < *targetPoolSize; i++ {
		go keyWriter(keyChan, targetPool, nil, wg)
	}

	wg.Wait()
	s.Stop()

}

func keyReader(c redis.Conn) (chan *redisKey, error) {
	keyChan := make(chan *redisKey, keyReaderBufferSize)

	// Scan keys into channel
	go func(c redis.Conn, keyChan chan *redisKey) {
		keys, err := redis.Strings(c.Do("KEYS", "*"))
		if err != nil {
			//return err
		}

		for _, k := range keys {
			rk, err := dumpKey(c, k)
			if err != nil {
				// do something
			}

			keyChan <- rk
		}

		close(keyChan)
	}(c, keyChan)

	return keyChan, nil
}

func keyWriter(k chan *redisKey, p *redis.Pool, errChan chan error, wg *sync.WaitGroup) {
	conn := p.Get()
	for k := range k {
		err := restoreKey(conn, k, *replaceKeys)
		if err != nil {
			log.Println("Error:", err)
		}
	}
	conn.Close()
	wg.Done()
}
