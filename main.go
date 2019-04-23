package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	APIKey    string `json:"api_key"`
	AuthToken string
	BaseURL   *url.URL

	httpClient *http.Client
}

func (c *Client) loadAPIKey() {
	apiKey, exists := os.LookupEnv("API_KEY")
	if exists == true {
		c.APIKey = apiKey
	} else {
		file, err := os.Open("config.json")
		if err != nil {
			log.Fatalf("Could not open config file: %v\n", err)
		}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(c)
		if err != nil {
			log.Fatalf("Could not decode config file: %v\n", err)
		}
	}
}

func NewClient(baseURL string, http *http.Client) *Client {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Could not parse Base URL: %v\n", err)
	}

	c := &Client{
		BaseURL: base,
	}
	c.loadAPIKey()

	return c
}

func main() {
	http := &http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	client := NewClient("https://au-api.basiq.io", http)

	fmt.Printf("Config is: %+v", client)
}
