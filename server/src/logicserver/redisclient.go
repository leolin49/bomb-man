package main

import (
	"paopao/server-base/src/base/env"

	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

type RedisManager struct {
	conn redis.Conn
}

var mRedisMgr *RedisManager

func RedisManager_GetMe() *RedisManager {
	if mRedisMgr == nil {
		c, err := redis.Dial("tcp", env.Get("redis", "server"))
		if err != nil {
			glog.Errorln("[redis] redis连接失败, ", err)
		}
		mRedisMgr = &RedisManager{
			conn: c,
		}
	}
	return mRedisMgr
}

func (this *RedisManager) Set(key, value string) {
	_, err := this.conn.Do("SET", key, value)
	if err != nil {
		glog.Errorln("[Redis] Set Error: ", err)
		return
	}
}

func (this *RedisManager) Get(key string) string {
	val, err := redis.String(this.conn.Do("GET", key))
	if err != nil {
		glog.Errorln("[Redis] Get Error: ", err)
		return ""
	}
	return val
}

func (this *RedisManager) HMSet(key string, fields interface{}) {
	_, err := this.conn.Do("HMSET", redis.Args{key}.AddFlat(fields)...)
	if err != nil {
		glog.Errorln("[Redis] HMSet Error: ", err)
		return
	}
}

func (this *RedisManager) HGet(key, field string) string {
	val, err := redis.String(this.conn.Do("HGET", key, field))
	if err != nil {
		glog.Errorln("[Redis] HGet Error: ", err)
		return ""
	}
	return val
}

func (this *RedisManager) HGetAll(key string) map[string]string {
	m, err := redis.StringMap(this.conn.Do("HGETALL", key))
	if err != nil {
		glog.Errorln("[Redis] HGetAll Error: ", err)
		return nil
	}
	return m
}

func (this *RedisManager) Exist(key string) bool {
	exist, err := redis.Bool(this.conn.Do("EXISTS", key))
	if err != nil {
		glog.Errorln("[Redis] EXISTS Error: ", err)
	}
	return exist
}

func (this *RedisManager) Stop() {
	err := this.conn.Close()
	if err != nil {
		glog.Errorln("[Redis] Close Error: ", err)
	}
}
