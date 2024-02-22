package main

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Authenticated  *bool           `json:"authenticated"`
	NumberOfFailed *int            `json:"number_of_failed"`
	ServerPort     *int            `json:"server_port"`
	Username       *string         `json:"username"`
	PassKey        *string         `json:"pass_key"`
	ProxyUrl       *[]ProxyUrlType `json:"proxy_url"`
	Redis          *RedisConfig    `json:"redis"`
}

type ProxyUrlType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type RedisConfig struct {
	Host string
	Port int
	Auth string
	Db   int
}

var config Config
var passkey string
var redisClient *redis.Client
var redisConn *redis.Conn

const REDIS_TIMEOUT = 5 * time.Minute
