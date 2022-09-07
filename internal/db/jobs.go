package db

import (
	"github.com/ercross/grabjobs/internal/models"
	"strings"
)

func (d *DB) SearchJobsByTitle(title string) ([]models.Job, error) {
	lowercasedTitle := strings.ToLower(title)
	return d.titleJobs[lowercasedTitle], nil
}

func (d *DB) FindJobsNearby(center models.Location, radius float32) ([]models.Job, error) {
	jobs := make([]models.Job, 0)
	return jobs, nil
}

func (d *DB) SearchJobsByTitleAndLocation(title string, location models.Location) ([]models.Job, error) {
	jobs := make([]models.Job, 0)
	return jobs, nil
}
