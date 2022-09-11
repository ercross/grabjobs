package db

import (
	"github.com/ercross/grabjobs/internal/models"
	"strings"
)

func (d *DB) TitleJobs() (map[string][]models.Job, error) {
	d.lock.RLock()
	titleJobs := d.titleJobs
	d.lock.RUnlock()
	return titleJobs, nil
}

func (d *DB) FindJobsNearby(center models.Location, radius float64) ([]models.Job, error) {
	d.lock.RLock()
	jobs := d.index.FindJobs(models.Distance{
		Unit:  models.Kilometer,
		Value: radius,
	}, center, d.titleJobs)
	d.lock.RUnlock()
	return jobs, nil
}

func (d *DB) SearchJobsByTitleAndLocation(title string, location models.Location) ([]models.Job, error) {
	d.lock.RLock()
	jobs := d.index.FindJobs(models.Distance{
		Unit:  models.Kilometer,
		Value: 5,
	}, location, d.titleJobs)
	d.lock.RUnlock()
	titleJobs := make([]models.Job, 0)
	for _, job := range jobs {
		if strings.ToLower(job.Title) == strings.ToLower(title) {
			titleJobs = append(titleJobs, job)
		}
	}
	return titleJobs, nil
}
