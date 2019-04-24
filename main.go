package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

type Transaction struct {
	Amount   string   `json:"amount"`
	SubClass SubClass `json:"subClass"`
}

type SubClass struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type Response struct {
	Type  string        `json:"type"`
	Count int64         `json:"count"`
	Size  int64         `json:"size"`
	Data  []Transaction `json:"data"`
}

type Status struct {
	Title             string
	NumOfTransactions int
	Sum               float64
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

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		log.Fatalf("Parsing URL failed: %v\n", err)
		return nil, err
	}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			log.Fatalf("Decoding Body failed: %v\n", err)
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		log.Fatalf("Could not create New Request: %v\n", err)
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)

	return resp, err
}

func (c *Client) CreateUser() (User, error) {
	fmt.Println("Creating a new User... gilfoyle (⌐■_■)")
	user := User{
		Email:  "gilfoyle@ppipper.com",
		Mobile: "+61410999666",
	}
	req, err := c.newRequest("POST", "/users", user)
	if err != nil {
		return User{}, err
	}

	var resUser User
	_, err = c.do(req, &resUser)
	if err != nil {
		return User{}, err
	}

	return resUser, nil
}

func (c *Client) Connect(userId string) (Job, error) {
	fmt.Println("Connecting and waiting for a new Job...")
	connection := Connection{
		LoginID:  "gavinBelson",
		Password: "hooli2016",
		Institution: Institution{
			ID: "AU00000",
		},
	}
	req, err := c.newRequest("POST", "/users/"+userId+"/connections", connection)
	if err != nil {
		return Job{}, err
	}

	var job Job
	_, err = c.do(req, &job)
	if err != nil {
		return Job{}, err
	}

	return job, nil
}

func (c *Client) CheckOnJob(jobId string) (string, error) {
	fmt.Print("Waiting...")
	req, err := c.newRequest("GET", "/jobs/"+jobId, nil)
	if err != nil {
		return "", err
	}

	// check in every 3 seconds
	for range time.Tick(time.Second * 3) {
		var job Job
		_, err = c.do(req, &job)
		if err != nil {
			return "", err
		}
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

	// if we got here probably error happened
	return "", errors.New("something went wrong")
}

func (c *Client) GetTransactions(path string) ([]Transaction, error) {
	fmt.Println("Fetching Transactions...")
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var transactionResponse Response
	_, err = c.do(req, &transactionResponse)
	if err != nil {
		return nil, err
	}

	return transactionResponse.Data, err
}

func (s *Status) AddTransaction(transaction Transaction) {
	s.NumOfTransactions++
	s.Title = transaction.SubClass.Title

	amount, _ := strconv.ParseFloat(transaction.Amount, 64)
	s.Sum += (math.Abs(amount))
}

func (s *Status) PrintAverage() {
	average := s.Sum / float64(s.NumOfTransactions)

	fmt.Println("-----------------------------------------------------")
	fmt.Printf("Category: %v\n", s.Title)
	fmt.Printf("Number of Transactions: %v\n", s.NumOfTransactions)
	fmt.Printf("Average Amount: %.3f\n", average)
}

func mapTransactions(transactions []Transaction) map[string]*Status {
	m := make(map[string]*Status)

	for _, transaction := range transactions {
		// we don't know where those without the code belong to
		// skip them
		code := transaction.SubClass.Code
		if code == "" {
			continue
		}

		storedStat, ok := m[code]
		if ok {
			storedStat.AddTransaction(transaction)
		} else {
			m[code] = &Status{}
			m[code].AddTransaction(transaction)
		}
	}

	return m
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

func printAverage(m map[string]*Status) {
	fmt.Println()
	fmt.Println("Hey, big spender:")

	for _, status := range m {
		status.PrintAverage()
	}
}

func main() {
	http := &http.Client{
		Timeout: time.Second * 5,
	}
	client := NewClient("https://au-api.basiq.io", "2.0", http)
	user, _ := client.CreateUser()
	job, _ := client.Connect(user.ID)
	transactionsLink, _ := client.CheckOnJob(job.ID)
	transactions, _ := client.GetTransactions(transactionsLink)
	mappedTransactions := mapTransactions(transactions)

	printAverage(mappedTransactions)
}
