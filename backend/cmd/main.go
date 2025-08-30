package main

import (
	"fmt"

	"github.com/osi-oss/osi/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Println(cfg)
}
