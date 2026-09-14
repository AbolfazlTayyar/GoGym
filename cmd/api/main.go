package main

import (
	"log"

	"github.com/AbolfazlTayyar/gogym/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("config loaded successfully: %s", cfg)
}
