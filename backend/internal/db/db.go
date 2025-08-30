package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/osi-oss/osi/internal/config"
)

func Connect(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.PgUser, cfg.PgPassword, cfg.PgDb, cfg.PgHost, cfg.PgPort)
	return openConnection(dsn)
}

func ConnectWithParams(user, password, dbname, host, port string) *sql.DB {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port)
	return openConnection(dsn)
}

func openConnection(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to db:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}

	return db
}
