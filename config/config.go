package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port    string
	BaseURL string
	CodeLen int
}

// Load reads configuration from environment variables with safe defaults.
func Load() *Config {
	port := getEnv("PORT", "8080")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)
	codeLen := clamp(parseInt(getEnv("CODE_LENGTH", "7")), 4, 20)

	return &Config{
		Port:    port,
		BaseURL: baseURL,
		CodeLen: codeLen,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
