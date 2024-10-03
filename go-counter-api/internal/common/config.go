package common

import (
	"os"
)

type Config struct {
	Host     string
	Pass     string
	Secret   string
	KafkaUrl string
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

	secret, found := os.LookupEnv("SECRET")
	if !found {
		panic("SECRET not set")
	}

	url, found := os.LookupEnv("KAFKA_URL")
	if !found {
		panic("KAFKA_URL not set")
	}

	return &Config{
		Host:     host,
		Pass:     pass,
		Secret:   secret,
		KafkaUrl: url,
	}
}