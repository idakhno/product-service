package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string        `env:"ENV" env-default:"local"`
	DatabaseURL string        `env:"DATABASE_URL" env-required:"true"`
	SentryDSN   string        `env:"SENTRY_DSN"`
	JWTSecret   string        `env:"JWT_SECRET" env-required:"true"`
	JWTTTL      time.Duration `env:"JWT_TTL" env-default:"24h"`
	HTTPServer
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:":8080"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file, relying on system environment variables: %v", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read config from environment variables: %v", err)
	}

	return &cfg
}
