package main

import (
	"context"
	"log"

	"siakad-mini/config"
	"siakad-mini/database"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	log.Println("terhubung ke database")
}
