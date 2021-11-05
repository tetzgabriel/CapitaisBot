package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

type country struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`
	Capital []string `json:"capital"`
}

func getClient(creds *Credentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}

	log.Printf("User's ACCOUNT: %s\n", user.Name)
	return client, nil
}

func main() {
	log.Printf("Starting CapitaisBot...")

	log.Printf("Getting Credentials from environment...")
	creds := Credentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	log.Printf("Success!\n")

	url := "https://restcountries.com/v3.1/all"

	countryclient := http.Client{
		Timeout: time.Second * 2,
	}

	req, countryerr := http.NewRequest(http.MethodGet, url, nil)
	if countryerr != nil {
		log.Fatal(countryerr)
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")

	res, getErr := countryclient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	fmt.Println(res)

	countries := []country{}
	jsonErr := json.Unmarshal(body, &countries)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	fmt.Println(countries[0])

	client, err := getClient(&creds)
	if err != nil {
		log.Println("Error getting Twitter Client")
		log.Println(err)
	}

	tweet, resp, err := client.Statuses.Update("A Test Tweet from a new Bot I'm building!", nil)

	if resp.StatusCode == 200 && err != nil {
		log.Printf("Success tweeting!")
		log.Printf("%+v\n", tweet)
	} else {
		log.Printf("Error tweeting!")
		log.Println(err)
	}
}
