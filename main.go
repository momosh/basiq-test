package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

type Job struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Steps []Step `json:"steps,omitempty"`
}

type Step struct {
	Title  string `json:"title"`
	Status string `json:"status"`
	Result Result `json:"result,omitempty"`
}

type Result struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func (j *Job) findStepIndexByTitle(title string) (int, error) {
	for i := range j.Steps {
		if j.Steps[i].Title == title {
			return i, nil
		}
	}

	return -1, errors.New("found nothing here")
}

func (c *Client) loadAPIKey() {
	fmt.Println("Checking for API key...")
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
	fmt.Println("Fetching Access Token...")
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
	fmt.Println("Creating a new User... gilfoyle (⌐■_■)")
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
		return User{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Creating new User failed: %v\n", err)
		return User{}, err
	}
	defer res.Body.Close()

	var resUser User
	err = json.NewDecoder(res.Body).Decode(&resUser)

	return resUser, err
}

func (c *Client) Connect(userId string) (Job, error) {
	fmt.Println("Connecting and waiting for a new Job...")
	rel := &url.URL{Path: "/users/" + userId + "/connections"}
	u := c.BaseURL.ResolveReference(rel)

	user := Connection{
		LoginID:  "gavinBelson",
		Password: "hooli2016",
		Institution: Institution{
			ID: "AU00000",
		},
	}
	data, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Could not create New Request: %v\n", err)
		return Job{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Creating new Connection failed: %v\n", err)
		return Job{}, err
	}
	defer res.Body.Close()

	var job Job
	err = json.NewDecoder(res.Body).Decode(&job)

	return job, err
}

func (c *Client) CheckOnJob(jobId string) (string, error) {
	fmt.Print("Waiting...")
	rel := &url.URL{Path: "/jobs/" + jobId}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Could not create New Request: %v\n", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// check every 5 seconds
	for range time.Tick(time.Second * 5) {
		res, err := c.httpClient.Do(req)
		if err != nil {
			log.Fatalf("Creating new Connection failed: %v\n", err)
			return "", err
		}
		defer res.Body.Close()

		var job Job
		err = json.NewDecoder(res.Body).Decode(&job)
		index, _ := job.findStepIndexByTitle("retrieve-transactions")
		// Job finished, return link
		if job.Steps[index].Status == "success" {
			fmt.Println(" Got the job!")
			return job.Steps[index].Result.URL, nil
		}
		if job.Steps[index].Status == "failed" {
			return job.Steps[index].Result.URL, errors.New("transaction job failed on server")
		}
		fmt.Print("..")
	}

	// if we got here probably
	return "", err
}

func NewClient(baseURL string, apiVersion string, http *http.Client) *Client {
	fmt.Println("Client initializing...")
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
	job, _ := client.Connect(user.ID)
	transactionsLink, _ := client.CheckOnJob(job.ID)

	fmt.Printf("TRANSACTIONS: %+v\n", transactionsLink)
}
