package main

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/otiai10/opengraph"
)

type project struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	ImageURL    string `json:"image"`
	Tweets      int    `json:"tweets"`
}

func (s *server) statisticBackgroundJob() {
	log.Printf("[INFO] Statistics job started")

	// Define some values for the statistics
	limit := 100
	cutoffDuration := 24 * time.Hour
	cutoffHour := roundToUnixHour(time.Now().Add(-cutoffDuration))

	// Delete all old entries in the database
	deleteOldSQL := `DELETE FROM tweets WHERE hour < ?;`
	statement, err := s.db.Prepare(deleteOldSQL)
	if err != nil {
		log.Printf("[ERROR] Unable to prepare delete old repos statement: %v", err.Error())
		return
	}
	defer statement.Close()
	res, err := statement.Exec(cutoffHour)
	if err != nil {
		log.Printf("[ERROR] Unable to delete old repos: %v", err.Error())
		return
	}
	if outdatedRows, err := res.RowsAffected(); err == nil {
		log.Printf("[INFO] Deleted %v outdated rows", outdatedRows)
	}

	// Query the database
	querySQL := `SELECT url, SUM(number) FROM tweets 
	GROUP BY url 
	ORDER BY SUM(number) DESC 
	LIMIT ?;`
	rows, err := s.db.Query(querySQL, limit)
	if err != nil {
		log.Printf("[ERROR] Unable query database for statitic: %v", err.Error())
		return
	}
	defer rows.Close()

	projects := make([]project, 0, limit)
	for rows.Next() {
		p := project{}
		rows.Scan(&p.URL, &p.Tweets)
		tmp, err := url.Parse(p.URL)
		if err != nil {
			continue
		}
		tmp.Scheme = "https"
		p.URL = tmp.String()

		projects = append(projects, p)
	}

	// Load metadata from GitHub
	wg := sync.WaitGroup{}
	wg.Add(len(projects))
	for i := range projects {
		go func(p *project) {
			defer wg.Done()
			addProjectMetadata(p)
		}(&projects[i])
	}
	wg.Wait()

	// Convert the projects to json and save it
	data, err := json.Marshal(projects)
	if err != nil {
		log.Printf("[ERROR] Unable to convert data to JSON: %v", err.Error())
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.projectsJSON = data

	log.Printf("[INFO] Statistics job finished")
}

func addProjectMetadata(p *project) {
	og, err := opengraph.Fetch(p.URL)
	if err != nil {
		log.Printf("[WARNING] Unable to load metadata for %v: %v", p.URL, err.Error())
		return
	}

	p.Description = og.Description
	p.Name = og.Title
	if len(og.Image) > 0 {
		p.ImageURL = og.Image[0].URL
	}
}
