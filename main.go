package main

import (
	"flag"
	"os"

	"github.com/kharljhon14/starbloom-server/cmd/api"
	"github.com/kharljhon14/starbloom-server/internal/jsonlog"
)

func main() {
	var cfg api.Config

	// Read the value of the port and env command-line flag intp the config struct
	flag.IntVar(&cfg.Port, "port", 8080, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Enviroment(development|staging|production)")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	app := &api.Application{
		Config: cfg,
		Logger: logger,
	}

	mux := app.Mount()

	err := app.Serve(mux)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}
