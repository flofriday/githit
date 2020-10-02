package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	_ "github.com/mattn/go-sqlite3"
)

var (
	consumerKey       = mustGetenv("TWITTER_API_KEY")
	consumerSecret    = mustGetenv("TWITTER_API_SECRET")
	accessToken       = mustGetenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret = mustGetenv("TWITTER_ACCESS_TOKEN_SECRET")
)

func mustGetenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("Missing required environment variable " + name)
	}
	return v
}

func setupDB(db *sql.DB) {
	//createTableSQL := `CREATE TABLE [IF NOT EXISTS] tweets (
	createTableSQL := `CREATE TABLE IF NOT EXISTS tweets (
		url TEXT,
		hour INTEGER,
		number INTEGER,
		PRIMARY KEY (url, hour)
	);`

	statement, err := db.Prepare(createTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer statement.Close()

	statement.Exec()
}

func main() {
	// Setup the database
	db, err := sql.Open("sqlite3", "data/githit.db")
	if err != nil {
		panic("Could not open databse: " + err.Error())
	}
	defer db.Close()
	setupDB(db)

	// Create the server
	s := server{
		db:           db,
		projectsJSON: []byte("{[]}"),
	}
	s.routes()

	// Start the background jobs
	go s.twitterBackgroundJob()
	go s.statisticBackgroundJob()
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(10).Minutes().Do(s.statisticBackgroundJob)
	scheduler.StartAsync()

	// Start listening
	addr := "0.0.0.0:3000"
	log.Printf("Server started at: %v", addr)
	http.ListenAndServe(addr, &s)
}
