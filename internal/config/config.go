package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Security SecurityConfig
	CORS     CORSConfig
	Logging  LoggingConfig
	WhatsApp WhatsAppConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  string
	Debug bool
}

type DatabaseConfig struct {
	Driver       string
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	SQLitePath   string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type SecurityConfig struct {
	APIKey          string
	JWTSecret       string
	RateLimitReqs   int
	RateLimitWindow time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type WhatsAppConfig struct {
	QRTimeout         time.Duration
	ReconnectInterval time.Duration
}

// Load carga la configuración desde variables de entorno
func Load() (*Config, error) {
	// Intentar cargar .env.local primero, luego .env
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := &Config{
		App: AppConfig{
			Name:  getEnv("APP_NAME", "kero-kero"),
			Env:   getEnv("APP_ENV", "development"),
			Port:  getEnv("APP_PORT", "8080"),
			Debug: getEnvBool("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			Driver:       getEnv("DB_DRIVER", "sqlite"),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", ""),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "kerokero_db"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			SQLitePath:   getEnv("SQLITE_PATH", "./data/kerokero.db"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		Security: SecurityConfig{
			APIKey:          getEnv("API_KEY", ""),
			JWTSecret:       getEnv("JWT_SECRET", ""),
			RateLimitReqs:   getEnvInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow: time.Duration(getEnvInt("RATE_LIMIT_WINDOW", 60)) * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "*"), ","),
			AllowedMethods: strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
			AllowedHeaders: strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"), ","),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		WhatsApp: WhatsAppConfig{
			QRTimeout:         time.Duration(getEnvInt("WA_QR_TIMEOUT", 60)) * time.Second,
			ReconnectInterval: time.Duration(getEnvInt("WA_RECONNECT_INTERVAL", 5)) * time.Second,
		},
	}

	// Validar configuración crítica
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate valida que la configuración sea correcta
func (c *Config) Validate() error {
	if c.Security.APIKey == "" && c.App.Env == "production" {
		return fmt.Errorf("API_KEY is required in production")
	}

	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("DB_DRIVER must be 'sqlite' or 'postgres', got: %s", c.Database.Driver)
	}

	if c.Database.Driver == "postgres" {
		if c.Database.Host == "" || c.Database.User == "" || c.Database.Name == "" {
			return fmt.Errorf("PostgreSQL requires DB_HOST, DB_USER, and DB_NAME")
		}
	}

	return nil
}

// GetDSN retorna el DSN para la base de datos
func (c *Config) GetDSN() string {
	if c.Database.Driver == "sqlite" {
		return fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", c.Database.SQLitePath)
	}

	// PostgreSQL
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr retorna la dirección de Redis
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// Helper functions

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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
