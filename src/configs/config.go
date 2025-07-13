package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var App Config

func LoadConfig(configPath string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(&App); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if App.App.Environment == "development" {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Println("Config file changed:", e.Name)
			if err := v.Unmarshal(&App); err != nil {
				log.Printf("Error reloading config: %v\n", err)
			}
		})
	}

	return &App, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "Brevity")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)

	v.SetDefault("database.sqlite.path", "./data/brevity.db")
	v.SetDefault("database.sqlite.busy_timeout", 5000)
	v.SetDefault("database.sqlite.foreign_keys", true)
	v.SetDefault("database.sqlite.journal_mode", "WAL")
	v.SetDefault("database.sqlite.cache_size", -2000)

	v.SetDefault("jwt.access_token_expiry", "15m")
	v.SetDefault("jwt.refresh_token_expiry", "168h")
	v.SetDefault("jwt.issuer", "brevity-service")
	v.SetDefault("jwt.secure_cookie", false)

	v.SetDefault("logger.level", "debug")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.file_path", "./logs/brevity.log")

	v.SetDefault("cors.enabled", true)
	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.max_age", "12h")

	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests", 100)
	v.SetDefault("rate_limit.window", "1m")
}

func GetConfigPath() string {
	paths := []string{
		"configs/app.yaml",
		"../configs/app.yaml",
		filepath.Join("src", "configs", "app.yaml"),
		"./app.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "configs/app.yaml"
}