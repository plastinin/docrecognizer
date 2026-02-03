package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	S3       S3Config
	Ollama   OllamaConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER" envDefault:"docrecognizer"`
	Password        string        `env:"DB_PASSWORD" envDefault:"secret"`
	Name            string        `env:"DB_NAME" envDefault:"docrecognizer"`
	SSLMode         string        `env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns        int           `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns        int           `env:"DB_MIN_CONNS" envDefault:"2"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"1h"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type S3Config struct {
	Endpoint  string `env:"S3_ENDPOINT" envDefault:"localhost:9000"`
	AccessKey string `env:"S3_ACCESS_KEY" envDefault:"minioadmin"`
	SecretKey string `env:"S3_SECRET_KEY" envDefault:"minioadmin"`
	Bucket    string `env:"S3_BUCKET" envDefault:"documents"`
	UseSSL    bool   `env:"S3_USE_SSL" envDefault:"false"`
}

type OllamaConfig struct {
	Host           string        `env:"OLLAMA_HOST" envDefault:"http://localhost:11434"`
	Model          string        `env:"OLLAMA_MODEL" envDefault:"qwen3-vl"`
	RequestTimeout time.Duration `env:"OLLAMA_REQUEST_TIMEOUT" envDefault:"5m"`
}

type LogConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"`
	// json или console
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Пытаемся загрузить .env файл (игнорируем ошибку, если файла нет)
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
