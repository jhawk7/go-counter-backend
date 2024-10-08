package common

import (
	"os"
)

type Config struct {
	RedisHost  string
	RedisPass  string
	KafkaHost  string
	KafkaTopic string
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

	khost, found := os.LookupEnv("KAFKA_HOST")
	if !found {
		panic("KAFKA_HOST not set")
	}

	topic, found := os.LookupEnv("KAFKA_TOPIC")
	if !found {
		panic("KAFKA_TOPIC not set")
	}

	return &Config{
		RedisHost:  host,
		RedisPass:  pass,
		KafkaHost:  khost,
		KafkaTopic: topic,
	}
}
