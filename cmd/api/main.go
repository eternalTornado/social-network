package main

import (
	"github.com/joho/godotenv"
	"github.com/truongndv/social/internal/db"
	"github.com/truongndv/social/internal/env"
	"github.com/truongndv/social/internal/store"
	"go.uber.org/zap"
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

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(cfg.db.addr, cfg.db.maxIdleConns, cfg.db.maxOpenConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
