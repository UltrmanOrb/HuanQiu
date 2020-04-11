package common

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)
//redis链接
func NewRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}

//set
func SetRedis(k string,v string) bool{
	newredis := NewRedisPool()
	conn := newredis.Get()
	defer conn.Close()
	_,err := conn.Do("SET",k,v)
	if err != nil {
		fmt.Println("redis get error",err)
		return false
	}
	return  true
}

//get
func GetRedis(k string) string {
	var val string
	newredis := NewRedisPool()
	conn := newredis.Get()
	defer conn.Close()
	val,err := redis.String(conn.Do("GET", k))
	if err != nil {
		fmt.Println("redis get error",err)

	}
	return val
}