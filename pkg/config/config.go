// Package config provides centralized configuration management for the Fern Platform
// It follows the twelve-factor app methodology for configuration management
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete platform configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Services   ServicesConfig   `mapstructure:"services"`
	Redis      RedisConfig      `mapstructure:"redis"`
	LLM        LLMConfig        `mapstructure:"llm"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout     time.Duration `mapstructure:"idleTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
	TLS             TLSConfig     `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"certFile"`
	KeyFile  string `mapstructure:"keyFile"`
}

type DatabaseConfig struct {
	Uri             string        `mapstructure:"uri"` // Deprecated, use individual fields instead
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	TimeZone        string        `mapstructure:"timezone"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

type AuthConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	JWTSecret     string        `mapstructure:"jwtSecret"`
	JWKSUrl       string        `mapstructure:"jwksUrl"`
	Issuer        string        `mapstructure:"issuer"`
	Audience      string        `mapstructure:"audience"`
	TokenExpiry   time.Duration `mapstructure:"tokenExpiry"`
	RefreshExpiry time.Duration `mapstructure:"refreshExpiry"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Structured bool   `mapstructure:"structured"`
}

type ServicesConfig struct {
	Reporter ServiceEndpoint `mapstructure:"reporter"`
	Mycelium ServiceEndpoint `mapstructure:"mycelium"`
	UI       ServiceEndpoint `mapstructure:"ui"`
}

type ServiceEndpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	URL  string `mapstructure:"url"`
}

type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Password    string        `mapstructure:"password"`
	DB          int           `mapstructure:"db"`
	PoolSize    int           `mapstructure:"poolSize"`
	IdleTimeout time.Duration `mapstructure:"idleTimeout"`
}

type LLMConfig struct {
	DefaultProvider string                 `mapstructure:"defaultProvider"`
	Providers       map[string]LLMProvider `mapstructure:"providers"`
	CacheEnabled    bool                   `mapstructure:"cacheEnabled"`
	CacheTTL        time.Duration          `mapstructure:"cacheTTL"`
	MaxTokens       int                    `mapstructure:"maxTokens"`
	Temperature     float32                `mapstructure:"temperature"`
}

type LLMProvider struct {
	Type    string `mapstructure:"type"`
	APIKey  string `mapstructure:"apiKey"`
	BaseURL string `mapstructure:"baseUrl"`
	Model   string `mapstructure:"model"`
	Enabled bool   `mapstructure:"enabled"`
}

type MonitoringConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
	Health  HealthConfig  `mapstructure:"health"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"serviceName"`
	Endpoint    string `mapstructure:"endpoint"`
}

type HealthConfig struct {
	Path     string        `mapstructure:"path"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

var globalConfig *Config

// Manager handles configuration initialization and management
type Manager struct {
	config *Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{}
}

// Load loads configuration from multiple sources with precedence:
// 1. Environment variables (highest)
// 2. Config file
// 3. Default values (lowest)
func (m *Manager) Load(configPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./config")
		viper.AddConfigPath("../config")
		viper.AddConfigPath("../../config")
		viper.AddConfigPath("/etc/fern-platform")
	}

	// Set defaults
	m.setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; continue with defaults and env vars
	}

	// Environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	m.bindEnvVars()

	// Unmarshal into config struct
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := m.validate(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.config = config
	globalConfig = config
	return nil
}

func (m *Manager) setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.readTimeout", "30s")
	viper.SetDefault("server.writeTimeout", "30s")
	viper.SetDefault("server.idleTimeout", "120s")
	viper.SetDefault("server.shutdownTimeout", "15s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.dbname", "fern_platform")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.timezone", "UTC")
	viper.SetDefault("database.maxOpenConns", 25)
	viper.SetDefault("database.maxIdleConns", 5)
	viper.SetDefault("database.connMaxLifetime", "300s")
	viper.SetDefault("database.connMaxIdleTime", "300s")

	// Auth defaults
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.tokenExpiry", "24h")
	viper.SetDefault("auth.refreshExpiry", "168h")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.structured", true)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.poolSize", 10)
	viper.SetDefault("redis.idleTimeout", "300s")

	// LLM defaults
	viper.SetDefault("llm.defaultProvider", "anthropic")
	viper.SetDefault("llm.cacheEnabled", true)
	viper.SetDefault("llm.cacheTTL", "1h")
	viper.SetDefault("llm.maxTokens", 4000)
	viper.SetDefault("llm.temperature", 0.7)

	// Monitoring defaults
	viper.SetDefault("monitoring.metrics.enabled", true)
	viper.SetDefault("monitoring.metrics.path", "/metrics")
	viper.SetDefault("monitoring.metrics.port", 9090)
	viper.SetDefault("monitoring.health.path", "/health")
	viper.SetDefault("monitoring.health.interval", "30s")
	viper.SetDefault("monitoring.health.timeout", "5s")
}

func (m *Manager) bindEnvVars() {
	// Server
	viper.BindEnv("server.port", "PORT", "SERVER_PORT")
	viper.BindEnv("server.host", "HOST", "SERVER_HOST")

	// Database
	viper.BindEnv("database.host", "DB_HOST", "POSTGRES_HOST")
	viper.BindEnv("database.port", "DB_PORT", "POSTGRES_PORT")
	viper.BindEnv("database.user", "DB_USER", "POSTGRES_USER")
	viper.BindEnv("database.password", "DB_PASSWORD", "POSTGRES_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME", "POSTGRES_DB")
	viper.BindEnv("database.sslmode", "DB_SSLMODE", "POSTGRES_SSLMODE")

	// Auth
	viper.BindEnv("auth.enabled", "AUTH_ENABLED")
	viper.BindEnv("auth.jwtSecret", "JWT_SECRET")
	viper.BindEnv("auth.jwksUrl", "JWKS_URL")
	viper.BindEnv("auth.issuer", "AUTH_ISSUER")
	viper.BindEnv("auth.audience", "AUTH_AUDIENCE")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// LLM Providers
	viper.BindEnv("llm.providers.anthropic.apiKey", "ANTHROPIC_API_KEY")
	viper.BindEnv("llm.providers.openai.apiKey", "OPENAI_API_KEY")
	viper.BindEnv("llm.providers.huggingface.apiKey", "HUGGINGFACE_API_KEY")

	// Logging
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
}

func (m *Manager) validate(config *Config) error {
	// Database validation
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// Auth validation
	if config.Auth.Enabled && config.Auth.JWTSecret == "" && config.Auth.JWKSUrl == "" {
		return fmt.Errorf("auth is enabled but no JWT secret or JWKS URL provided")
	}

	return nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if globalConfig == nil {
		panic("configuration not initialized - call config.Load() first")
	}
	return globalConfig
}

// GetDatabaseConfig returns database configuration
func GetDatabaseConfig() *DatabaseConfig {
	return &GetConfig().Database
}

// GetServerConfig returns server configuration
func GetServerConfig() *ServerConfig {
	return &GetConfig().Server
}

// GetAuthConfig returns authentication configuration
func GetAuthConfig() *AuthConfig {
	return &GetConfig().Auth
}

// GetServicesConfig returns services configuration
func GetServicesConfig() *ServicesConfig {
	return &GetConfig().Services
}

// GetLLMConfig returns LLM configuration
func GetLLMConfig() *LLMConfig {
	return &GetConfig().LLM
}

// DatabaseConnectionString returns the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s.fern-platform:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// MigrationURL returns the PostgreSQL URL for migrations
func (d *DatabaseConfig) MigrationURL() string {
	dbString := fmt.Sprintf(
		"postgres://%s:%s@%s.fern-platform:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
	fmt.Printf("db string is : [%s]\n", dbString)
	return dbString
}

// RedisConnectionString returns the Redis connection string
func (r *RedisConfig) ConnectionString() string {
	if r.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d", r.Password, r.Host, r.Port, r.DB)
	}
	return fmt.Sprintf("redis://%s:%d/%d", r.Host, r.Port, r.DB)
}
