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

var sourceRedis = flag.String("source", "redis://127.0.0.1:6379", "source Redis host (default localhost:6379)")
var sourceDB = flag.Int("source-database", 0, "source Redis DB (default 0)")
var targetRedis = flag.String("target", "", "target Redis database")
var targetDB = flag.Int("target-database", 0, "target Redis DB (default 0)")
var targetPoolSize = flag.Int("write-concurrency", 10, "target Redis write concurrency")
var replaceKeys = flag.Bool("replace", false, "replace existing keys on target")

func main() {
	flag.Parse()

	// Connect to source
	srcPool := newPoolFromURL(*sourceRedis, *targetPoolSize+1)
	srcRedis := srcPool.Get()

	// Select database
	if *sourceDB != 0 {
		srcRedis.Do("SELECT", *sourceDB)
	}

	// Error Channel
	errChan := make(chan error)
	go func(ec chan error) {
		for err := range ec {
			log.Println("Error:", err)
		}
	}(errChan)

	// Read Keys
	keyChan, err := keyReader(srcRedis)
	if err != nil {
		log.Fatal(err)
	}

	valueReaderWG := &sync.WaitGroup{}

	// Read Values
	rkeyChan := make(chan *redisKey, keyReaderBufferSize)
	for i := 0; i < *targetPoolSize; i++ {
		valueReaderWG.Add(1)
		go valueReader(keyChan, rkeyChan, srcPool, errChan, valueReaderWG)
	}

	// Connect to target
	targetPool := newPoolFromURL(*targetRedis, *targetPoolSize)

	// Setup spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Syncing Redis Databases "
	s.Start()

	writerWG := &sync.WaitGroup{}
	for i := 0; i < *targetPoolSize; i++ {
		writerWG.Add(1)
		go keyWriter(rkeyChan, targetPool, errChan, writerWG)
	}

	valueReaderWG.Wait()
	close(rkeyChan)
	writerWG.Wait()
	s.Stop()

	close(errChan)

}

func keyReader(c redis.Conn) (chan string, error) {
	keyChan := make(chan string, keyReaderBufferSize)

	// Scan keys into channel
	go func(c redis.Conn, keyChan chan string) {
		keys, err := redis.Strings(c.Do("KEYS", "*"))
		if err != nil {
			//return err
			log.Println("key read error:", err)
		}

		for _, k := range keys {
			keyChan <- k
		}

		close(keyChan)
		c.Close()
	}(c, keyChan)

	return keyChan, nil
}

func valueReader(keys chan string, rkeys chan *redisKey, p *redis.Pool, errChan chan error, wg *sync.WaitGroup) {
	conn := p.Get()
	if *sourceDB != 0 {
		conn.Do("SELECT", *sourceDB)
	}
	for k := range keys {
		rk, err := dumpKey(conn, k)
		if err != nil {
			errChan <- err
		}

		rkeys <- rk
	}
	conn.Close()
	wg.Done()

}

func keyWriter(keys chan *redisKey, p *redis.Pool, errChan chan error, wg *sync.WaitGroup) {
	conn := p.Get()
	for k := range keys {
		err := restoreKey(conn, k, *replaceKeys)
		if err != nil {
			log.Println("Error:", err)
		}
	}
	conn.Close()
	wg.Done()
}
