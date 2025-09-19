package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"log/slog"
)

var logger *slog.Logger

func init() {
	logger = NewLogger()
}
func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}

}
