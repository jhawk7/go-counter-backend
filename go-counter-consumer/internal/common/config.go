package common

import (
	"os"
)

type Config struct {
	Host string
	Pass string
}

func GetConfig() *Config {
	host, found := os.LookupEnv("REDIS_HOST")
	if !found {
		panic("REDIS_HOST not set")
	}

	pass, found := os.LookupEnv("REDIS_PASS")
	if !found {
		panic("REDIS_PASS not set")
	}

	return &Config{
		Host: host,
		Pass: pass,
	}
}
