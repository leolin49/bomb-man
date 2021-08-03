package main

import (
	"paopao/server-base/src/base/env"
	"time"

	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

type RedisManager struct {
	pool *redis.Pool
}

var redisMgr RedisManager

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,               // 池中空闲连接的最大数量。
		MaxActive:   10,               // 池在给定时间分配的最大连接数。
		IdleTimeout: 60 * time.Second, // 连接保持空闲的最大时间
		Wait:        true,             // 达到最大连接数后，是否等待，Get()方法将会阻塞
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				glog.Errorln("[Redis Connect] Connect Failed ", err)
				return nil, err
			}
			pwd := env.Get("redis", "password")
			if len(pwd) != 0 {
				if _, err := c.Do("AUTH", pwd); err != nil {
					glog.Errorln("[Redis] password error")
					c.Close()
					return nil, err
				}
			}
			db := env.Get("redis", "db")
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		// 在将连接返回给应用程序之前，使用 TestOnBorrow 函数检查空闲连接的运行状况。
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (this *RedisManager) NewRedisManager() bool {
	server := env.Get("redis", "server")
	this.pool = newPool(server)
	if this.pool == nil {
		glog.Errorln("[Redis] connect error")
		return false
	}
	return true
}

func (this *RedisManager) Set(key, value string) {
	conn := this.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	if err != nil {
		glog.Errorln("[Redis] Set Error: ", err)
		return
	}
}

func (this *RedisManager) Get(key string) string {
	conn := this.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		glog.Errorln("[Redis] Get Error: ", err)
		return ""
	}
	return val
}

func (this *RedisManager) HMSet(key string, fields interface{}) {
	conn := this.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", redis.Args{key}.AddFlat(fields)...)
	if err != nil {
		glog.Errorln("[Redis] HMSet Error: ", err)
		return
	}
}

func (this *RedisManager) HGet(key, field string) string {
	conn := this.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("HGET", key, field))
	if err != nil {
		glog.Errorln("[Redis] HGet Error: ", err)
		return ""
	}
	return val
}

func (this *RedisManager) HGetAll(key string) map[string]string {
	conn := this.pool.Get()
	defer conn.Close()
	m, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		glog.Errorln("[Redis] HGetAll Error: ", err)
		return nil
	}
	return m
}

func (this *RedisManager) Exist(key string) bool {
	conn := this.pool.Get()
	defer conn.Close()
	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		glog.Errorln("[Redis] EXISTS Error: ", err)
	}
	return exist
}
