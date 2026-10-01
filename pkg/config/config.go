package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds settings common to every service, loaded via Viper.
// Each service can embed this and add its own fields (DB URLs, etc.).
type Config struct {
	ServiceName string // ECOM_SERVICE_NAME
	Env         string // ECOM_ENV  (dev|staging|prod)
	HTTPPort    int    // ECOM_HTTP_PORT
	LogLevel    string // ECOM_LOG_LEVEL (debug|info|warn|error)
}

// Load reads configuration with precedence: env vars > config file > defaults.
func Load() (*Config, error) {
	v := viper.New()

	// defaults
	v.SetDefault("service_name", "service")
	v.SetDefault("env", "dev")
	v.SetDefault("http_port", 8080)
	v.SetDefault("log_level", "info")

	// environment variables: key http_port -> ECOM_HTTP_PORT
	v.SetEnvPrefix("ECOM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// optional config file (merged if found, ignored if absent)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{
		ServiceName: v.GetString("service_name"),
		Env:         v.GetString("env"),
		HTTPPort:    v.GetInt("http_port"),
		LogLevel:    v.GetString("log_level"),
	}
	return cfg, nil
}

// IsProd reports whether we are running in a production environment.
func (c *Config) IsProd() bool {
	return c.Env == "prod"
}
