package main

import (
	"flag"
	"log"
	"os"

	"github.com/kharljhon14/starbloom-server/cmd/api"
)

const version = "1.0.0"

func main() {
	var cfg api.Config

	// Read the value of the port and env command-line flag intp the config struct
	flag.IntVar(&cfg.Port, "port", 8080, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Enviroment(development|staging|production)")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &api.Application{
		Config: cfg,
		Logger: logger,
	}

	mux := app.Mount()

	logger.Fatalln(app.Serve(mux))

}
