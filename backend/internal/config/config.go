package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	IsProd       bool
	AppName      string
	AppHost      string
	AppPort      int
	FiberPrefork bool

	DBHost            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBPort            int
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime int
}

// Load reads .env (if present, for local dev) and then real environment
// variables, which always win — in Kubernetes config comes from
// Secret/ConfigMap env vars, never from a .env baked into the image.
func Load() (*Config, error) {
	for _, path := range []string{"./", "../../"} {
		viper.SetConfigFile(path + ".env")
		if err := viper.ReadInConfig(); err == nil {
			break
		}
	}

	viper.AutomaticEnv()

	viper.SetDefault("APP_HOST", "0.0.0.0")
	viper.SetDefault("APP_PORT", 8080)
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 60)

	cfg := &Config{
		IsProd:       viper.GetString("APP_ENV") == "prod",
		AppName:      viper.GetString("APP_NAME"),
		AppHost:      viper.GetString("APP_HOST"),
		AppPort:      viper.GetInt("APP_PORT"),
		FiberPrefork: viper.GetBool("FIBER_PREFORK"),

		DBHost:            viper.GetString("DB_HOST"),
		DBUser:            viper.GetString("DB_USER"),
		DBPassword:        viper.GetString("DB_PASSWORD"),
		DBName:            viper.GetString("DB_NAME"),
		DBPort:            viper.GetInt("DB_PORT"),
		DBMaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		DBMaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		DBConnMaxLifetime: viper.GetInt("DB_CONN_MAX_LIFETIME"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate fails fast at startup listing EVERY missing variable at once.
func (c *Config) Validate() error {
	required := []struct {
		env   string
		value string
	}{
		{"APP_NAME", c.AppName},
		{"DB_HOST", c.DBHost},
		{"DB_USER", c.DBUser},
		{"DB_PASSWORD", c.DBPassword},
		{"DB_NAME", c.DBName},
	}

	var missing []string
	for _, r := range required {
		if strings.TrimSpace(r.value) == "" {
			missing = append(missing, r.env)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("[CONFIG] validation failed. missing variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
