package main

import (
	"log"
	"net/url"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

func (s *server) twitterBackgroundJob() {
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessTokenSecret, consumerKey, consumerSecret)

	stream := api.PublicStreamFilter(url.Values{
		"track":    []string{"github com"},
		"language": []string{"en"},
	})

	defer stream.Stop()

	log.Print("[Info] Started Twitter streaming...\n")
	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Printf("received unexpected value of type %T", v)
			continue
		}

		for _, url := range t.Entities.Urls {
			link := githubRegex.FindString(url.Expanded_url)
			if link == "" {
				log.Printf("[No match] @%v: %v", t.User.ScreenName, url.Expanded_url)
				continue
			}

			log.Printf("@%v: %v (%v)", t.User.ScreenName, link, url.Expanded_url)
			s.addGitHubRepo(link)
		}
	}
}

func (s *server) addGitHubRepo(repo string) {
	hour := (time.Now().UTC().UnixNano() / time.Hour.Nanoseconds()) * time.Hour.Nanoseconds()
	hour /= time.Second.Nanoseconds()

	addRepoSQL := `INSERT INTO tweets VALUES(?, ?, 1)
	ON CONFLICT(url, hour) DO UPDATE SET number = number + 1;`
	statement, err := s.db.Prepare(addRepoSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer statement.Close()

	_, err = statement.Exec(repo, hour)
	if err != nil {
		log.Printf("[ERROR] Unable to add repo to sqlite: %v", err.Error())
	}
}
