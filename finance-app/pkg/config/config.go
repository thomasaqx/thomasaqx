package config

import (
	"os"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	GRPC     GRPCConfig
}

type ServerConfig struct {
	Port string
}

type PostgresConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers string
}

type GRPCConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			DSN: getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=financeapp port=5432 sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Kafka: KafkaConfig{
			Brokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "9090"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
