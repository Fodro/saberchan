package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/database/psql"
	"github.com/Fodro/saberchan/internal/health"
	"github.com/Fodro/saberchan/internal/server"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

func main() {
	log.Print("starting server")
	conf := config.ParseConfig()
	log.Print("env parsed")

	var repo database.Repository
	if conf.DB.DBType == "postgres" {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			conf.DB.DBHost, conf.DB.DBPort, conf.DB.DBUser, conf.DB.DBPass, conf.DB.DBName, conf.DB.SSLMode)
		log.Print("running migrations")
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
			return
		}
		goose.SetDialect("postgres")
		err = goose.Up(db, "./migrations")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Print("succeded running migrations")
		repo = psql.NewRepo(connStr)
	} else {
		log.Fatalf("unsupported db type: %s", conf.DB.DBType)
		return
	}

	board := board.NewService(repo, conf)
	health := health.NewService(repo)
	server := server.NewServer(conf, board, health)
	log.Println("starting server on port", conf.Port)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("shutting down server")

		if err := server.Stop(context.Background()); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}()

	if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
}
