package v1

import (
	"fmt"
	"github.com/ercross/grabjobs/internal/models"
	"log"
	"net/http"
	"time"
)

type repository interface {
	// TitleJobs fetches a mapping of title to available jobs
	TitleJobs() (map[string][]models.Job, error)

	// FindJobsNearby finds jobs within radius of location.
	// If radius is zero, a default radius of 5km is used.
	// FindJobsNearby returns an empty slice if no job is found within radius of location.
	// Any error returned is an internal error
	FindJobsNearby(location models.Location, radius float64) ([]models.Job, error)

	// SearchJobsByTitleAndLocation finds jobs matching title within 5km
	// radius of location.
	// If no job is found matching title within the default radius,
	// an empty slice of Models.Job is returned.
	// Any error returned is an internal error
	SearchJobsByTitleAndLocation(title string, location models.Location) ([]models.Job, error)
}

type Config struct {
	LocationDataFilePath string
	Port                 int
}

type App struct {
	repo   repository
	Routes http.Handler
	Config Config
}

func (app *App) StartServer() error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.Port),
		Handler:           app.Routes,
		ReadHeaderTimeout: 40 * time.Second,
		WriteTimeout:      60 * time.Second,
		MaxHeaderBytes:    2048,
	}
	log.Printf("Server started and listening on port %d", app.Config.Port)
	return server.ListenAndServe()
}
