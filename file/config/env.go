package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Port    string  `env:"PORT" envDefault:"8080"`
	Storage string `env:"STORAGE" envDefault:"s3"`

	DB *Database
	S3 *S3
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

type S3 struct {
	AccessKey string `env:"S3_ACCESS_KEY"`
	SecretKey string `env:"S3_SECRET_KEY"`
	Bucket    string `env:"S3_BUCKET"`
	Region    string `env:"S3_REGION"`
	Url 	 string `env:"S3_URL"`
	EnableExpriration bool `env:"S3_ENABLE_EXPIRATION" envDefault:"false"`
	FileExpire time.Duration `env:"S3_FILE_EXPIRE" envDefault:"3600s"`
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
	s3 := &S3{}
	if err := env.Parse(s3); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	cfg.S3 = s3
	return cfg
}
