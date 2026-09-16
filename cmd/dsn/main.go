// Command dsn prints the PostgreSQL connection string built from the app's
// config, for use by the Makefile's migrate-up / migrate-down targets so the
// connection details stay defined in one place (internal/config).
package main

import (
	"fmt"

	"github.com/AbolfazlTayyar/gogym/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Print(cfg.DB.DSN())
}
