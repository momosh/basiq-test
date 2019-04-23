package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	APIKey      string `json:"api_key"`
	AccessToken string `json:"access_token"`
	APIv        string
	BaseURL     *url.URL

	httpClient *http.Client
}

type Institution struct {
	ID string
}

type Connection struct {
	LoginID     string
	Password    string
	Institution Institution
}

type User struct {
	ID     string `json:"id,omitempty"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
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

func (c *Client) getAuthToken() {
	rel := &url.URL{Path: "/token"}
	u := c.BaseURL.ResolveReference(rel)

	data := url.Values{}
	data.Set("scope", "SERVER_ACCESS")

	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatalf("Could not create New Request: %v\n", err)
	}
	req.Header.Set("Authorization", "Basic "+c.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("basiq-version", c.APIv)

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Getting Auth token failed: %v\n", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(c)
	if err != nil {
		log.Fatalf("Decoding Body failed: %v\n", err)
	}
}

func (c *Client) CreateUser() (User, error) {
	rel := &url.URL{Path: "/users"}
	u := c.BaseURL.ResolveReference(rel)

	user := User{
		Email:  "gilfoyle@ppipper.com",
		Mobile: "+61410999666",
	}
	data, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Could not create New Request: %v\n", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Creating new User failed: %v\n", err)
	}
	defer res.Body.Close()

	var resUser User
	err = json.NewDecoder(res.Body).Decode(&resUser)

	return resUser, err
}

func NewClient(baseURL string, apiVersion string, http *http.Client) *Client {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Could not parse Base URL: %v\n", err)
	}

	c := &Client{
		BaseURL:    base,
		APIv:       apiVersion,
		httpClient: http,
	}
	c.loadAPIKey()
	c.getAuthToken()

	return c
}

func main() {
	http := &http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	client := NewClient("https://au-api.basiq.io", "2.0", http)
	user, _ := client.CreateUser()

	fmt.Printf("User ID: %v\n", user.ID)
	fmt.Printf("User Email: %v\n", user.Email)
}
