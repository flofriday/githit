package main

import (
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

var (
	gitURLRegex       = regexp.MustCompile(`^(https?:\/\/)?(www.)?(github|gitlab).com\/(?!(sponsors|settings|api))[a-zA-Z0-9\-_]+\/[a-zA-Z0-9\-_\.]+`)
	gitForbiddenRegex = regexp.MustCompile(`^(https?:\/\/)?(www.)?(github|gitlab).com\/(sponsors|settings|api)`)
	gitRepoRegex      = regexp.MustCompile(`(github|gitlab).com\/[a-zA-Z0-9\-_]+\/[a-zA-Z0-9\-_\.]+`)
)

func (s *server) twitterBackgroundJob() {
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessTokenSecret, consumerKey, consumerSecret)

	stream := api.PublicStreamFilter(url.Values{
		"track":    []string{"github com", "gitlab com"},
		"language": []string{"en"},
	})

	defer stream.Stop()

	log.Print("[INFO] Started Twitter streaming...\n")
	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Printf("received unexpected value of type %T", v)
			continue
		}

		for _, url := range t.Entities.Urls {
			expandedURL := strings.ToLower(url.Expanded_url)
			link := gitURLRegex.FindString(expandedURL)
			if link == "" {
				log.Printf("[NO MATCH] @%v: %v", t.User.ScreenName, url.Expanded_url)
				continue
			}

			if gitForbiddenRegex.MatchString(expandedURL) {
				log.Printf("[NO MATCH] @%v: %v", t.User.ScreenName, url.Expanded_url)
				continue
			}

			log.Printf("[MATCH] @%v: %v", t.User.ScreenName, url.Expanded_url)
			s.addGitHubRepo(link)
		}
	}
}

func (s *server) addGitHubRepo(repo string) {
	// Format the URL so all entries are formatted the same
	// This means the url will have: no scheme, no user or password, no query
	// no www in front of the host, no queries, no fragment
	repo = gitRepoRegex.FindString(repo)
	if repo == "" {
		log.Printf("[ERROR] Unable to parse repo url")
		return
	}

	// Perpare the sql statement
	addRepoSQL := `INSERT INTO tweets VALUES(?, ?, 1)
	ON CONFLICT(url, hour) DO UPDATE SET number = number + 1;`
	statement, err := s.db.Prepare(addRepoSQL)
	if err != nil {
		log.Printf("[ERROR] Unable to prepare add repo statement: %v", err.Error())
		return
	}
	defer statement.Close()

	// Execute the sql statement
	hour := roundToUnixHour(time.Now().UTC())
	_, err = statement.Exec(repo, hour)
	if err != nil {
		log.Printf("[ERROR] Unable to add repo to sqlite: %v", err.Error())
		return
	}
}
