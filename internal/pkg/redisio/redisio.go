package redisio

import (
	"context"

	"github.com/garyburd/redigo/redis"
)

type RedisIO struct {
	pool *redis.Pool
	conn redis.Conn
}

type NewInput struct {
	RedisURL string
}

func New(input NewInput) (*RedisIO, error) {
	redispool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", input.RedisURL)
		},
	}
	// Get a connection
	conn := redispool.Get()
	defer conn.Close()
	// Test the connection
	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}
	return &RedisIO{
		conn: conn,
		pool: redispool,
	}, nil
}

func (r *RedisIO) Publish(ctx context.Context, key string, msg []byte) error {
	conn := r.pool.Get()
	conn.Do("PUBLISH", key, msg)
	return nil
}

func (r *RedisIO) Subscribe(ctx context.Context, key string, msg chan []byte) error {
	rc := r.pool.Get()
	psc := redis.PubSubConn{Conn: rc}
	if err := psc.PSubscribe(key); err != nil {
		return err
	}

	go func() {
		for {
			switch v := psc.Receive().(type) {
			case redis.PMessage:
				msg <- v.Data
			}
		}
	}()
	return nil
}
