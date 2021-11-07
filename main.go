package main

import (
	"bufio"
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
	executeBot()
}

func executeBot() {
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
		log.Fatal("--- Error getting Twitter Client, shutting down app :( --- ")
		log.Println(err)
	}

	isCountryAlreadyTweeted := true
	countryToTweet := Country{}

	for {
		countryToTweet = getRandomCountry(countries)

		isCountryAlreadyTweeted = stringInSlice(countryToTweet.Name.Common, countries)

		if !isCountryAlreadyTweeted {
			break
		} else {
			log.Printf("Country already tweeted, getting another one ;)\n")
		}
	}

	clearRequestData(res)
	tweet(client, &countryToTweet)

	writeCountryNameInFile(err, countryToTweet)
}

func stringInSlice(a string, list []Country) bool {
	for _, b := range list {
		if b.Name.Common == a {
			return false
		}
	}
	return true
}

func writeCountryNameInFile(err error, countryToTweet Country) {
	f, fileErr := os.OpenFile("data/TweetedCountries.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if fileErr != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(countryToTweet.Name.Common + "\n"); err != nil {
		panic(err)
	}
}

func readCountriesFromFile() ([]string, error) {
	file, err := os.Open("data/TweetedCountries.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
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
