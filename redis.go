package main

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

func newPoolFromURL(URL string, maxActive int) *redis.Pool {
	return &redis.Pool{
		MaxActive:   maxActive,
		MaxIdle:     maxActive,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(URL) },
	}
}

type redisSummary struct {
	Version         string
	Keys            int
	OpsPerSecond    int
	UsedMemoryHuman string
}

func fetchSummary(c redis.Conn) (*redisSummary, error) {
	return &redisSummary{}, nil
}

func dumpKey(c redis.Conn, key string) (*redisKey, error) {
	value, err := redis.String(c.Do("DUMP", key))
	if err != nil {
		return nil, err
	}
	return &redisKey{key, value, 0}, nil
}

func restoreKey(c redis.Conn, rk *redisKey, replace bool) error {
	if replace {
		_, err := c.Do("RESTORE", rk.Key, rk.TTL, rk.Value, "REPLACE")
		return err
	}
	_, err := c.Do("RESTORE", rk.Key, rk.TTL, rk.Value)
	return err
}

func getTTL(c redis.Conn, key string) (int, error) {
	return redis.Int(c.Do("PTTL", key))
}
