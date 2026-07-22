package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port          string        `env:"PORT" envDefault:"8080"`
	StoreIp       bool          `env:"STORE_IP" envDefault:"false"`
	Secret        string        `env:"SECRET" envDefault:"dontuseme"`
	AdminAPIToken string        `env:"ADMIN_API_TOKEN"`
	PurgeInterval time.Duration `env:"PURGE_INTERVAL" envDefault:"10m"`
	// Comma-separated CIDRs (or IPs) allowed to set X-Forwarded-For.
	// Empty = never trust XFF (use RemoteAddr only).
	TrustedProxies string `env:"TRUSTED_PROXIES"`

	Redis *Redis
	DB    *Database
	S3    *S3
}

type Database struct {
	DBType  string `env:"DB_TYPE"`
	DBHost  string `env:"DB_HOST"`
	DBPort  int    `env:"DB_PORT"`
	DBUser  string `env:"DB_USER"`
	DBPass  string `env:"DB_PASS"`
	DBName  string `env:"DB_NAME"`
	SSLMode string `env:"SSL_MODE"`

	MaxConns        int32         `env:"DB_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"DB_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m"`
	MaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"5m"`
}

type Redis struct {
	Host     string        `env:"REDIS_HOST"`
	Port     string        `env:"REDIS_PORT"`
	User     string        `env:"REDIS_USER"`
	Password string        `env:"REDIS_PASS"`
	Expires  time.Duration `env:"REDIS_EXPIRES" envDefault:"1m"`
}

type S3 struct {
	AccessKey         string        `env:"S3_ACCESS_KEY"`
	SecretKey         string        `env:"S3_SECRET_KEY"`
	Bucket            string        `env:"S3_BUCKET"`
	Region            string        `env:"S3_REGION"`
	Url               string        `env:"S3_URL"`
	PublicURL         string        `env:"S3_PUBLIC_URL"` // browser origin, e.g. https://example.com (links become {PublicURL}/{bucket}/{key})
	UseSSL            bool          `env:"S3_USE_SSL" envDefault:"true"`
	ForcePathStyle    bool          `env:"S3_FORCE_PATH_STYLE" envDefault:"false"`
	EnableExpriration bool          `env:"S3_ENABLE_EXPIRATION" envDefault:"false"`
	FileExpire        time.Duration `env:"S3_FILE_EXPIRE" envDefault:"3600s"`
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
	redis := &Redis{}
	if err := env.Parse(redis); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	cfg.Redis = redis
	s3 := &S3{}
	if err := env.Parse(s3); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	cfg.S3 = s3
	return cfg
}
