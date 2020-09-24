package main

import (
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/ChimeraCoder/anaconda"
)

var (
	consumerKey       = mustGetenv("TWITTER_API_KEY")
	consumerSecret    = mustGetenv("TWITTER_API_SECRET")
	accessToken       = mustGetenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret = mustGetenv("TWITTER_ACCESS_TOKEN_SECRET")
	githubRegex       = regexp.MustCompile(`^(https?:\/\/)?(www.)?(github|gitlab).com\/[a-zA-Z0-9\-_]+\/[a-zA-Z0-9\-_]+`)
)

func mustGetenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("Missing required environment variable " + name)
	}
	return v
}

func main() {
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessTokenSecret, consumerKey, consumerSecret)

	stream := api.PublicStreamFilter(url.Values{
		"track":    []string{"github com"},
		"language": []string{"en"},
	})

	defer stream.Stop()

	log.Print("[Info] Started streaming...\n")
	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Printf("received unexpected value of type %T", v)
			continue
		}

		for _, url := range t.Entities.Urls {
			//link := url.Display_url
			link := githubRegex.FindString(url.Expanded_url)
			if link == "" {
				log.Printf("[No match] @%v: %v", t.User.ScreenName, url.Expanded_url)
				continue
			}
			log.Printf("@%v: %v (%v)", t.User.ScreenName, link, url.Expanded_url)
		}
	}
}
