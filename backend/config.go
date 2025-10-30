package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DBConf struct {
	DSN string `yaml:"dsn"`
}

type Config struct {
	Port         int    `yaml:"port"`
	JWTSecret    string `yaml:"jwt_secret"`
	AuthDisabled bool   `yaml:"auth_disabled"`
	CookieSecure bool   `yaml:"cookie_secure"`
	Database     DBConf `yaml:"database"`
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	// env overrides
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		c.CookieSecure = v == "1" || v == "true" || v == "TRUE"
	}
	if v := os.Getenv("AUTH_DISABLED"); v != "" {
		c.AuthDisabled = v == "1" || v == "true" || v == "TRUE"
	}
	if v := os.Getenv("PORT"); v != "" {
		c.Port = atoiDefault(v, c.Port)
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET/config.jwt_secret is required")
	}
	if c.Database.DSN == "" {
		return nil, fmt.Errorf("DB_DSN/config.database.dsn is required")
	}
	return &c, nil
}
