package db

import (
	"encoding/csv"
	"fmt"
	"github.com/ercross/grabjobs/internal/models"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type DB struct {
	lock *sync.RWMutex

	// titleJobs index jobs based on job titles.
	// This enables fast retrieval of jobs based on job titles.
	// Alternatively, this indexing could be done with any
	// standard geospatial based DBMS.
	titleJobs map[string][]models.Job
}

// Initialize initializes the DB.
// filepath is the path to the location.csv file.
func Initialize(filepath string) (*DB, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error: failed to open file on path %s: %v", filepath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error encountered reading file on path %s : %v", filepath, err)
	}

	db := loadTitleJobs(removeTitleLine(lines))
	db.lock = new(sync.RWMutex)
	return db, nil
}

// removeTitleLine removes title line if present in
// Some csv file may contain table titles on the first line.
// Remove first line in lines if it contains the table titles
func removeTitleLine(lines [][]string) [][]string {
	removeFirstLine := false
	firstLine := lines[0]

	// if second and/or third column isn't a valid float value,
	// then first line is title line
	_, err := strconv.ParseFloat(firstLine[1], 32)
	if err != nil {
		removeFirstLine = true
	}

	_, err = strconv.ParseFloat(firstLine[2], 32)
	if err != nil {
		removeFirstLine = true
	}

	if removeFirstLine {
		return lines[1:]
	}

	return lines
}

// loadTitleJobs reads job on each line of lines into DB.
// Each line in lines must contain job title, longitude, latitude
// in that order of indexing
func loadTitleJobs(lines [][]string) *DB {
	jobs := make(map[string][]models.Job)
	var db DB
	if len(lines) == 0 {
		return &db
	}

	for i, line := range lines {
		var job models.Job

		// check that line contains exactly 3 items,
		// else line is incomplete and skipped
		if len(line) != 3 {
			continue
		}

		longitude, err := strconv.ParseFloat(line[1], 32)
		if err != nil {
			log.Printf("error parsing longitude on line %d", i)
			continue
		}
		latitude, err := strconv.ParseFloat(line[2], 32)
		if err != nil {
			log.Printf("error parsing latitude on line %d", i)
			continue
		}
		job.Title = line[0]
		job.Location = models.Location{
			Longitude: longitude,
			Latitude:  latitude,
		}

		// check that map contains jobs with same title,
		// else initialize new slice for jobs with job.Title
		//
		// Map keys are converted to lower case to eliminate case sensitivity
		// when searching for jobs based on title.
		// Ensure also that job title search queries are converted
		// to lower case before using on titleJobs
		lowercasedTitle := strings.ToLower(job.Title)
		if sameJobs, ok := jobs[lowercasedTitle]; ok {
			sameJobs = append(sameJobs, job)
			jobs[lowercasedTitle] = sameJobs
		} else {
			jobs[lowercasedTitle] = []models.Job{job}
		}

	}

	db.titleJobs = jobs
	return &db
}
