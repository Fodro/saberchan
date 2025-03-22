package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/database/psql"
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

	fmt.Println(repo)
}
