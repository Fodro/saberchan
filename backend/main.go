package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/ban"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/captcha"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/database/psql"
	"github.com/Fodro/saberchan/internal/file/s3service"
	"github.com/Fodro/saberchan/internal/follow"
	"github.com/Fodro/saberchan/internal/health"
	"github.com/Fodro/saberchan/internal/purge"
	"github.com/Fodro/saberchan/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Print("starting server")
	conf := config.ParseConfig()
	log.Print("env parsed")

	var repo database.Repository
	var pool *pgxpool.Pool
	if conf.DB.DBType == "postgres" {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			conf.DB.DBHost, conf.DB.DBPort, conf.DB.DBUser, conf.DB.DBPass, conf.DB.DBName, conf.DB.SSLMode)

		ctx := context.Background()
		cfg, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			log.Fatal(err)
			return
		}
		cfg.MaxConns = conf.DB.MaxConns
		cfg.MinConns = conf.DB.MinConns
		cfg.MaxConnLifetime = conf.DB.MaxConnLifetime
		cfg.MaxConnIdleTime = conf.DB.MaxConnIdleTime

		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer pool.Close()

		log.Print("running migrations")
		db := stdlib.OpenDBFromPool(pool)
		if err := goose.SetDialect("postgres"); err != nil {
			panic(fmt.Errorf("goose dialect: %w", err))
		}
		if err := goose.Up(db, "./migrations"); err != nil {
			panic(fmt.Errorf("migrations: %w", err))
		}
		if err := db.Close(); err != nil {
			log.Printf("stdlib bridge close: %v", err)
		}
		log.Print("succeded running migrations")
		repo = psql.NewRepo(pool)
	} else {
		log.Fatalf("unsupported db type: %s", conf.DB.DBType)
		return
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host + ":" + conf.Redis.Port,
		Username: conf.Redis.User,
		Password: conf.Redis.Password,
	})

	captcha := captcha.NewService(redisClient, conf.Redis.Expires)
	ban := ban.NewService(ban.NewRedisStore(redisClient), repo)
	followSvc := follow.NewService(follow.NewRedisStore(redisClient), repo)

	file := s3service.NewService(conf)
	board := board.NewService(repo, file, conf, followSvc)
	health := health.NewServiceWithRedis(repo, redisClient)
	server := server.NewServer(conf, board, captcha, health, ban, followSvc)
	log.Println("starting server on port", conf.Port)

	purgeCtx, cancelPurge := context.WithCancel(context.Background())
	go purge.Run(purgeCtx, repo, file, conf.PurgeInterval)
	log.Printf("purge worker interval %s", conf.PurgeInterval)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("shutting down server")

		cancelPurge()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Stop(shutdownCtx); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}()

	if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
}
