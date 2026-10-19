package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/truongndv/social/internal/db"
	"github.com/truongndv/social/internal/env"
	"github.com/truongndv/social/internal/store"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	API cho mạng xã hội GopherSocial
//	@version		0.0.1
//	@BasePath		/v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {

	_ = godotenv.Load()

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgresql://admin@hotdog123/socialnetwork:8080?sslmode=false"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:    env.GetString("ENV", "development"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
	}

	db, err := db.New(cfg.db.addr, cfg.db.maxIdleConns, cfg.db.maxOpenConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
