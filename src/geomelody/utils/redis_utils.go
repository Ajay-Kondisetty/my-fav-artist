package utils

import (
	"errors"
	"fmt"
	"geomelody/constants"
	"github.com/gomodule/redigo/redis"
)

func RedisConn() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", constants.REDIS_HOST, constants.REDIS_PORT))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func RedisGetData(conn redis.Conn, key string) (string, error) {
	data, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", errors.New("failed to get data from Redis")
	}

	return data, nil
}

func RedisSetData(conn redis.Conn, key, val string, ttl int) (bool, error) {
	_, err := conn.Do("SET", key, val, "EX", ttl)
	if err != nil {
		return false, errors.New("failed to set data in Redis")
	}

	return true, nil
}
