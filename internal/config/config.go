package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config contains application configuration.
// All parameters are loaded from environment variables.
type Config struct {
	Env         string        `env:"ENV" env-default:"local"`                    // Environment: local, dev, prod
	DatabaseURL string        `env:"DATABASE_URL" env-required:"true"`           // PostgreSQL connection URL
	SentryDSN   string        `env:"SENTRY_DSN"`                                 // Sentry DSN (optional)
	JWTSecret   string        `env:"JWT_SECRET" env-required:"true"`             // Secret key for JWT token signing
	JWTTTL      time.Duration `env:"JWT_TTL" env-default:"24h"`                   // JWT token lifetime
	HTTPServer                                                                   // HTTP server settings
}

// HTTPServer contains HTTP server configuration.
type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:":8080"`     // Server address and port
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`       // Read/write timeout
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"` // Idle connection timeout
}

// MustLoad loads configuration from environment variables.
// First attempts to load .env file, then reads system environment variables.
// Terminates the program with an error if required parameters are not set.
func MustLoad() *Config {
	// Attempt to load .env file (not critical if it doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file, relying on system environment variables: %v", err)
	}

	var cfg Config

	// Read configuration from environment variables
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read config from environment variables: %v", err)
	}

	return &cfg
}
