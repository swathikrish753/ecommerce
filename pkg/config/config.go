package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName   string // ECOM_SERVICE_NAME
	Env           string // ECOM_ENV
	HTTPPort      int    // ECOM_HTTP_PORT
	LogLevel      string // ECOM_LOG_LEVEL
	DBDSN         string // ECOM_DB_DSN
	JWTSecret     string // ECOM_JWT_SECRET
	JWTTTLMinutes int    // ECOM_JWT_TTL_MINUTES
	MongoURI      string // ECOM_MONGO_URI
	MongoDB       string // ECOM_MONGO_DB
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("service_name", "service")
	v.SetDefault("env", "dev")
	v.SetDefault("http_port", 8080)
	v.SetDefault("log_level", "info")
	v.SetDefault("db_dsn", "postgres://ecom:ecom_pass@localhost:5432/ecom_auth?sslmode=disable")
	v.SetDefault("jwt_secret", "dev-secret-change-me")
	v.SetDefault("jwt_ttl_minutes", 60)
	v.SetDefault("mongo_uri", "mongodb://ecom:ecom_pass@localhost:27017")
	v.SetDefault("mongo_db", "ecom_catalog")

	v.SetEnvPrefix("ECOM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

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
		ServiceName:   v.GetString("service_name"),
		Env:           v.GetString("env"),
		HTTPPort:      v.GetInt("http_port"),
		LogLevel:      v.GetString("log_level"),
		DBDSN:         v.GetString("db_dsn"),
		JWTSecret:     v.GetString("jwt_secret"),
		JWTTTLMinutes: v.GetInt("jwt_ttl_minutes"),
		MongoURI:      v.GetString("mongo_uri"),
		MongoDB:       v.GetString("mongo_db"),
	}
	return cfg, nil
}

func (c *Config) IsProd() bool {
	return c.Env == "prod"
}
