package main

import (
	"flag"
	current "github.com/ercross/grabjobs/cmd/api/v1" // simply change import path if current api version changes
	"github.com/ercross/grabjobs/internal/db"
	"log"
)

func main() {
	app := new(current.App)
	app.Config = initConfig()
	repo, err := db.Initialize(app.Config.LocationDataFilePath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	app.Routes = current.Routes(repo)
	if err := app.StartServer(); err != nil {
		log.Fatalf("error encountered starting server: %v", err)
	}
}

func initConfig() current.Config {
	var config current.Config
	flag.IntVar(&config.Port, "port", 4042, "port the server listens on")
	flag.StringVar(&config.LocationDataFilePath, "db", "empty", "api endpoint")
	flag.Parse()
	return config
}
