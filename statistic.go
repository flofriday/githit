package main

import (
	"encoding/json"
	"log"
	"time"
)

type project struct {
	Name        string
	URL         string
	Description string
	ImageURL    string
	Stars       int
	Tweets      int
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
		projects = append(projects, p)
	}

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
