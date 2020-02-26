package redis

import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "time"
)

var(
  pool *redis.Pool
  redisHost = "127.0.0.1:6379"
  redisPass = ""
)

func newRedisPool() *redis.Pool {
  return &redis.Pool{
    Dial: func() (conn redis.Conn, err error) {
      c, err := redis.Dial("tcp", redisHost)
      if err != nil {
        fmt.Println(err)
        return nil, err
      }
      //if _, err = c.Do("AUTH", redisPass); err !=nil{
      //  c.Close()
      //  return nil, err
      //}
      return c, nil
    },
    TestOnBorrow: func(conn redis.Conn, t time.Time) error {
      if time.Since(t) < time.Minute {
        return nil
      }
      _, err := conn.Do("PING")
      return err
    },
    MaxIdle:      50,
    MaxActive:    30,
    IdleTimeout:  300*time.Second,
  }
}

func init() {
  pool = newRedisPool()
}

func RedisPool() *redis.Pool {
  return pool
}
