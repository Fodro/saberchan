package config

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	Port    int  `env:"PORT" envDefault:"8080"`
	StoreIp bool `env:"STORE_IP" envDefault:"false"`
	Secret string `env:"SECRET" envDefault:"dontuseme"`

	DB *Database
}

type Database struct {
	DBType  string `env:"DB_TYPE"`
	DBHost  string `env:"DB_HOST"`
	DBPort  int    `env:"DB_PORT"`
	DBUser  string `env:"DB_USER"`
	DBPass  string `env:"DB_PASS"`
	DBName  string `env:"DB_NAME"`
	SSLMode string `env:"SSL_MODE"`
}

func ParseConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	db := &Database{}
	if err := env.Parse(db); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	cfg.DB = db
	return cfg
}
