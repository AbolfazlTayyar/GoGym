// Command dsn prints the database DSN for the Makefile's migrate targets.
package main

import (
	"fmt"

	"github.com/AbolfazlTayyar/gogym/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Print(cfg.DB.DSN())
}
