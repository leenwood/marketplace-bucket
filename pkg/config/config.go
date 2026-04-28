// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	HTTP    HTTPConfig
	GRPC    GRPCConfig
	Redis   RedisConfig
	Otel    OtelConfig
	CartTTL time.Duration
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// GRPCConfig contains gRPC server settings.
type GRPCConfig struct {
	Addr string
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// OtelConfig contains OpenTelemetry settings.
type OtelConfig struct {
	Endpoint    string
	ServiceName string
}

// Load reads configuration from environment variables, falling back to defaults.
//
// Environment variable mapping (prefix CART_ for all except the special cases):
//
//	REDIS_ADDR                    → redis.addr
//	REDIS_PASSWORD                → redis.password
//	REDIS_DB                      → redis.db
//	OTEL_EXPORTER_OTLP_ENDPOINT   → otel.endpoint
//	CART_HTTP_ADDR                → http.addr
//	CART_GRPC_ADDR                → grpc.addr
//	CART_TTL                      → cart_ttl
func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("grpc.addr", ":9090")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("otel.endpoint", "localhost:4317")
	v.SetDefault("otel.service_name", "marketplace-bucket")
	v.SetDefault("cart_ttl", "168h") // 7 days

	v.SetEnvPrefix("CART")
	v.AutomaticEnv()

	_ = v.BindEnv("redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("otel.endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT")

	cfg := &Config{
		HTTP: HTTPConfig{
			Addr:         v.GetString("http.addr"),
			ReadTimeout:  v.GetDuration("http.read_timeout"),
			WriteTimeout: v.GetDuration("http.write_timeout"),
		},
		GRPC: GRPCConfig{
			Addr: v.GetString("grpc.addr"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("redis.addr"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		Otel: OtelConfig{
			Endpoint:    v.GetString("otel.endpoint"),
			ServiceName: v.GetString("otel.service_name"),
		},
		CartTTL: v.GetDuration("cart_ttl"),
	}

	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf("redis.addr must not be empty")
	}

	return cfg, nil
}
