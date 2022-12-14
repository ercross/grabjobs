package v1

import (
	"fmt"
	"github.com/ercross/grabjobs/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func Routes(repo repository) http.Handler {
	mux := chi.NewMux()
	app := new(App)
	app.repo = repo

	mux.MethodNotAllowed(app.sendMethodNotAllowedResponse)
	mux.NotFound(app.sendNotFoundResponse)

	mux.Route("/api/v1", func(r chi.Router) {
		r.Mount("/jobs", app.jobsRouter())
	})

	return mux
}

func (app *App) jobsRouter() chi.Router {
	router := chi.NewRouter()

	router.Get("/available", app.getTitleJobs)
	router.Get("/nearby", app.getJobsNearby)
	router.Get("/top-jobs/around-me", app.getTopTitleJobsAround)
	return router
}

// getTitleJobs fetches a mapping of title to available jobs
// Request Method: GET
// Query Parameters: None
// Response Type: application/json
func (app *App) getTitleJobs(w http.ResponseWriter, r *http.Request) {

	jobs, err := app.repo.TitleJobs()
	if err != nil {
		app.sendServerErrorResponse(w, fmt.Errorf("error searching jobs by title: %v", err))
		return
	}

	app.sendJSONResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Available jobs",
	}, jobs)
}

// getJobsNearby fetches jobs some radius around current location
// Request Method: GET
// Query Parameters:
//
//	latitude 	decimal/float
//	longitude 	decimal/float
//	radius 		decimal/float
//
// Response Type: application/json
func (app *App) getJobsNearby(w http.ResponseWriter, r *http.Request) {

	// read query paramters
	latitude, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 32)
	if err != nil {
		app.sendFailedValidationResponse(w, map[string]string{"latitude": "latitude not a valid decimal/float"})
		return
	}

	longitude, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 32)
	if err != nil {
		app.sendFailedValidationResponse(w, map[string]string{"longitude": "longitude not a valid decimal/float"})
		return
	}

	radius, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	if err != nil {
		app.sendFailedValidationResponse(w, map[string]string{"radius": "radius not a valid decimal/float"})
		return
	}

	jobs, err := app.repo.FindJobsNearby(models.Location{
		Longitude: longitude,
		Latitude:  latitude,
	}, radius)

	if err != nil {
		app.sendServerErrorResponse(w, fmt.Errorf("error encountered finding jobs within a radius of %f", radius))
		return
	}

	app.sendJSONResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Jobs around you",
	}, jobs)
}

// getTopTitleJobsAround fetches up to 5 jobs some arbitrary radius
// around current location matching the specified title.
// Request Method: GET
// Query Parameters:
//
//	latitude 	decimal/float
//	longitude 	decimal/float
//	title 		string
//
// Response Type: application/json
func (app *App) getTopTitleJobsAround(w http.ResponseWriter, r *http.Request) {

	// read query paramters
	latitude, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 32)
	if err != nil {
		app.sendFailedValidationResponse(w, map[string]string{"latitude": "latitude not a valid decimal/float"})
		return
	}

	longitude, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 32)
	if err != nil {
		app.sendFailedValidationResponse(w, map[string]string{"longitude": "longitude not a valid decimal/float"})
		return
	}

	title := r.URL.Query().Get("title")
	if notValidString(title) {
		app.sendFailedValidationResponse(w, map[string]string{"title": "title is not a valid text"})
		return
	}

	location := models.Location{
		Longitude: longitude,
		Latitude:  latitude,
	}
	jobs, err := app.repo.SearchJobsByTitleAndLocation(title, location)

	if err != nil {
		app.sendServerErrorResponse(w, fmt.Errorf("error encountered finding %v jobs around %v", title, location))
		return
	}

	app.sendJSONResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    fmt.Sprintf("Top %v Jobs around you", title),
	}, jobs)
}
