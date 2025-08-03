package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	LogLevel    string `mapstructure:"log_level"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type JWTConfig struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
}

func LoadConfig() (*Config, error) {
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	if err := loadEnvFile(env); err != nil {
		fmt.Printf("Warning: could not load .env file for %s environment: %v\n", env, err)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	configPaths := findConfigPaths()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	v.SetConfigName("base")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	if env != "base" {
		v.SetConfigName(env)
		if err := v.MergeInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to merge %s config: %w", env, err)
			}
		}
	}

	v.AutomaticEnv()
	bindEnvironmentVariables(v)

	v.Set("app.environment", env)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&config)

	return &config, nil
}

func loadEnvFile(env string) error {
	projectRoot := findProjectRootFromCwd()

	envFiles := []string{
		filepath.Join(projectRoot, fmt.Sprintf(".env.%s", env)),
		fmt.Sprintf(".env.%s", env),
		filepath.Join(projectRoot, ".env"),
		".env",
	}

	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			fmt.Printf("Loaded environment from %s\n", envFile)
			return nil
		}
	}

	if env == "test" {
		return nil
	}

	return fmt.Errorf("no .env file found")
}

func findConfigPaths() []string {
	var paths []string

	if projectRoot := findProjectRootFromCwd(); projectRoot != "" {
		paths = append(paths, filepath.Join(projectRoot, "config"))
	}

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Dir(filename)
		if projectRoot := findProjectRoot(dir); projectRoot != "" {
			configPath := filepath.Join(projectRoot, "config")
			if len(paths) == 0 || paths[0] != configPath {
				paths = append(paths, configPath)
			}
		}
	}

	fallbackPaths := []string{
		"config",
		"./config",
		"../config",
		"../../config",
		"../../../config",
	}

	for _, p := range fallbackPaths {
		paths = append(paths, p)
	}

	return paths
}

func findProjectRootFromCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return findProjectRoot(cwd)
}

func findProjectRoot(startPath string) string {
	dir := startPath
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func bindEnvironmentVariables(v *viper.Viper) {
	// Database bindings
	v.BindEnv("database.host", "DATABASE_HOST")
	v.BindEnv("database.port", "DATABASE_PORT")
	v.BindEnv("database.user", "DATABASE_USER")
	v.BindEnv("database.password", "DATABASE_PASSWORD")
	v.BindEnv("database.db_name", "DATABASE_NAME")
	v.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE")
	v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	v.BindEnv("database.conn_max_idle_time", "DATABASE_CONN_MAX_IDLE_TIME")

	// Server bindings
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	// JWT bindings
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.access_token_duration", "JWT_ACCESS_TOKEN_DURATION")
	v.BindEnv("jwt.refresh_token_duration", "JWT_REFRESH_TOKEN_DURATION")

	// Redis bindings
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")
	v.BindEnv("redis.dial_timeout", "REDIS_DIAL_TIMEOUT")
	v.BindEnv("redis.read_timeout", "REDIS_READ_TIMEOUT")
	v.BindEnv("redis.write_timeout", "REDIS_WRITE_TIMEOUT")
	v.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")

	// App bindings
	v.BindEnv("app.name", "APP_NAME")
	v.BindEnv("app.debug", "APP_DEBUG")
	v.BindEnv("app.log_level", "APP_LOG_LEVEL")
}

func setDefaults(config *Config) {
	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 15 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 15 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

	// Database defaults
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}
	if config.Database.ConnMaxIdleTime == 0 {
		config.Database.ConnMaxIdleTime = 1 * time.Minute
	}

	// JWT defaults
	if config.JWT.AccessTokenDuration == 0 {
		config.JWT.AccessTokenDuration = 15 * time.Minute
	}
	if config.JWT.RefreshTokenDuration == 0 {
		config.JWT.RefreshTokenDuration = 7 * 24 * time.Hour
	}

	// Redis defaults
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}
	if config.Redis.DB == 0 && config.App.Environment == "test" {
		config.Redis.DB = 1 // Use DB 1 for tests by default
	}
	if config.Redis.DialTimeout == 0 {
		config.Redis.DialTimeout = 5 * time.Second
	}
	if config.Redis.ReadTimeout == 0 {
		config.Redis.ReadTimeout = 3 * time.Second
	}
	if config.Redis.WriteTimeout == 0 {
		config.Redis.WriteTimeout = 3 * time.Second
	}
	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	// App defaults
	if config.App.Name == "" {
		config.App.Name = "ccards"
	}
	if config.App.LogLevel == "" {
		if config.App.Debug {
			config.App.LogLevel = "debug"
		} else {
			config.App.LogLevel = "info"
		}
	}
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production" || c.App.Environment == "prod"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == "dev"
}

func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}

func (c *Config) IsLocal() bool {
	return c.App.Environment == "local"
}
