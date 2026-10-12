package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/truongndv/social/internal/db"
	"github.com/truongndv/social/internal/env"
	"github.com/truongndv/social/internal/store"
)

func main() {
	_ = godotenv.Load()

	addr := env.GetString("DB_ADDR", "postgres://admin:hotdog123@localhost:5432/socialnetwork?sslmode=disable")

	conn, err := db.New(addr, 30, 30, "15m")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	store := store.NewStorage(conn)
	db.Seed(store)
}
