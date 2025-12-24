package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	PostgresURL       string
	KafkaBrokers      []string
	AllowedOrigins    []string
	BotWaitSeconds    int
	ReconnectSeconds  int
}

func Load() Config {
	return Config{
		Port:             getenv("PORT", "8080"),
		PostgresURL:      getenv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable"),
		KafkaBrokers:     split(getenv("KAFKA_BROKERS", "localhost:9092")),
		AllowedOrigins:   split(getenv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		BotWaitSeconds:   getenvInt("BOT_WAIT_SECONDS", 10),
		ReconnectSeconds: getenvInt("RECONNECT_SECONDS", 30),
	}
}

func split(val string) []string {
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
