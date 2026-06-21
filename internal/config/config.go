package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName             string
	AppPort             string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPass              string
	DBName              string
	JWTSecret           string
	JWTAccessExpiryMin  int
	JWTRefreshExpiryDay int
	RedisHost           string
	RedisPort           string
	RedisPassword       string
	UploadDir           string
	OpenAIKey           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:             getEnv("APP_NAME", "interior-backend"),
		AppPort:             getEnv("APP_PORT", "8080"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPass:              getEnv("DB_PASS", ""),
		DBName:              getEnv("DB_NAME", "postgres"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTAccessExpiryMin:  getEnvInt("JWT_ACCESS_EXPIRY_MIN", 60),
		JWTRefreshExpiryDay: getEnvInt("JWT_REFRESH_EXPIRY_DAY", 7),
		RedisHost:           getEnv("REDIS_HOST", "localhost"),
		RedisPort:           getEnv("REDIS_PORT", "6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		OpenAIKey:           getEnv("OPENAI_API_KEY", ""),
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return ":" + c.AppPort
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		c.DBHost,
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBPort,
	)
}

// DBDSNDefault connects to the default 'postgres' database, used for creating the actual database if it doesn't exist.
func (c *Config) DBDSNDefault() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable TimeZone=Asia/Jakarta",
		c.DBHost,
		c.DBUser,
		c.DBPass,
		c.DBPort,
	)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func (c *Config) AccessTokenDuration() time.Duration {
	return time.Duration(c.JWTAccessExpiryMin) * time.Minute
}

func (c *Config) RefreshTokenDuration() time.Duration {
	return time.Duration(c.JWTRefreshExpiryDay) * 24 * time.Hour
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := getEnv(key, "")
	if val == "" {
		return fallback
	}
	var result int
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		return fallback
	}
	return result
}
