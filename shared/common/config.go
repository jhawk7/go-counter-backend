package common

import (
	"os"
)

type Config struct {
	RedisHost  string
	RedisPass  string
	MQTTServer string
	MQTTPort   string
	MQTTUser   string
	MQTTPass   string
	MQTTTopic  string
	MQTTQos    string
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

	mqttServer, found := os.LookupEnv("MQTT_SERVER")
	if !found {
		panic("MQTT_SERVER not set")
	}

	mqttPort, found := os.LookupEnv("MQTT_PORT")
	if !found {
		panic("MQTT_PORT not set")
	}

	mqttUser, found := os.LookupEnv("MQTT_USER")
	if !found {
		panic("MQTT_USER not set")
	}

	mqttPass, found := os.LookupEnv("MQTT_PASS")
	if !found {
		panic("MQTT_PASS not set")
	}

	mqttTopic, found := os.LookupEnv("MQTT_TOPIC")
	if !found {
		panic("MQTT_TOPIC not set")
	}

	mqttQos, found := os.LookupEnv("MQTT_QOS")
	if !found {
		panic("MQTT_QOS not set")
	}

	return &Config{
		RedisHost:  host,
		RedisPass:  pass,
		MQTTServer: mqttServer,
		MQTTPort:   mqttPort,
		MQTTUser:   mqttUser,
		MQTTPass:   mqttPass,
		MQTTTopic:  mqttTopic,
		MQTTQos:    mqttQos,
	}
}
