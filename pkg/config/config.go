package config

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	// Add more configuration sections as needed
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port int
	// Add more server configuration fields as needed
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	// Add more database configuration fields as needed
}

// LoadConfig loads configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	// Implementation details
	return &Config{}, nil
}
