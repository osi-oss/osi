package main

import (
	"fmt"

	"github.com/osi-oss/osi/internal/config"
	"github.com/osi-oss/osi/internal/db"
)

func main() {
	cfg := config.Load()

	pgDb := db.Connect(&cfg)

	fmt.Println(pgDb.Stats())
}
