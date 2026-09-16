package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string
	JWTExpireHrs int
	HISBaseURL   string
}

func LoadConfig() *Config {
	expireHrs, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	if expireHrs <= 0 {
		expireHrs = 24
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "hospital_middleware"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    getEnv("JWT_SECRET", "super-secret-hospital-jwt-key-agnos-2026"),
		JWTExpireHrs: expireHrs,
		HISBaseURL:   getEnv("HIS_BASE_URL", "https://hospital-a.api.co.th"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
