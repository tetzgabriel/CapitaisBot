package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
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

type Country struct {
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

	log.Printf("Preparing request...\n")
	url := os.Getenv("COUNTRY_API_URL")
	countryclient, req := prepareRequest(url)

	log.Printf("Calling countries API...\n")
	res := getCountries(countryclient, req)

	log.Printf("Parsing Data...\n")
	countries := parseResponse(res)
	fmt.Println(countries[0])

	log.Printf("Getting Twitter client...\n")
	client, err := getClient(&creds)
	if err != nil {
		log.Println("Error getting Twitter Client")
		log.Println(err)
	}

	countryToTweet := getRandomCountry(countries)
	clearRequestData(res)
	tweet(client, &countryToTweet)
}

func getRandomCountry(countries []Country) Country {
	num := getRandomInt()

	return countries[num]
}

func getRandomInt() int {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := 248

	return rand.Intn(max-min+1) + min
}

func tweet(client *twitter.Client, country *Country) {
	tweet, _, err := client.Statuses.Update("Country: "+country.Name.Common+"  -  Capital City: "+country.Capital[0], nil)
	if err != nil {
		log.Printf("Error tweeting!")
		log.Fatal(err)
	}

	log.Printf("Success tweeting!")
	log.Printf("%+v\n", tweet)
}

func parseResponse(res *http.Response) []Country {
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	fmt.Println(res)

	countries := []Country{}
	jsonErr := json.Unmarshal(body, &countries)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return countries
}

func clearRequestData(res *http.Response) {
	if res.Body != nil {
		defer res.Body.Close()
	}
}

func prepareRequest(url string) (http.Client, *http.Request) {
	countryclient := http.Client{
		Timeout: time.Second * 2,
	}

	req, countryerr := http.NewRequest(http.MethodGet, url, nil)
	if countryerr != nil {
		log.Fatal(countryerr)
	}

	req.Header.Set("User-Agent", "capitaisBot")

	return countryclient, req
}

func getCountries(countryclient http.Client, req *http.Request) *http.Response {
	res, getErr := countryclient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	return res
}
