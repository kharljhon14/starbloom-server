package main

import (
	"context"
	"flag"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/kharljhon14/starbloom-server/cmd/api"
	"github.com/kharljhon14/starbloom-server/internal/data"
	"github.com/kharljhon14/starbloom-server/internal/jsonlog"
)

func main() {
	var cfg api.Config

	// Read the value of the port and env command-line flag intp the config struct
	flag.IntVar(&cfg.Port, "port", 8080, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Enviroment(development|staging|production)")
	flag.StringVar(&cfg.Db.Dsn, "DSN", os.Getenv("DSN"), "database connection string")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDb(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
		os.Exit(1)
	}

	logger.PrintInfo("database connection pool establised", nil)

	app := &api.Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
	}

	mux := app.Mount()

	err = app.Serve(mux)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDb(cfg api.Config) (*pgx.Conn, error) {
	db, err := pgx.Connect(context.Background(), cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
