package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	App     AppConfig
	Database DatabaseConfig
	Broker  BrokerConfig
}

type AppConfig struct {
	Port     string
	LogLevel string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type BrokerConfig struct {
	RabbitMQ RabbitMQConfig
	Kafka    KafkaConfig
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Queue    string
}

type KafkaConfig struct {
	Brokers string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Port:     getEnv("APP_PORT", "8080"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Broker: BrokerConfig{
			RabbitMQ: RabbitMQConfig{
				URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
				Exchange: getEnv("RABBITMQ_EXCHANGE", "fms"),
				Queue:    getEnv("RABBITMQ_QUEUE", "fms-consumers"),
			},
			Kafka: KafkaConfig{
				Brokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
			},
		},
	}

	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (c *Config) GetRabbitMQUrl() string {
	return c.Broker.RabbitMQ.URL
}

func (c *Config) GetKafkaBrokers() []string {
	return strings.Split(c.Broker.Kafka.Brokers, ",")
}