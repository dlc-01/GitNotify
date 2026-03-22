package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dlc-01/GitNotify/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	_, err := config.Load(*path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}
