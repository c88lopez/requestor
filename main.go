package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const coreDomain = "safiro.jampp.com"

var wg sync.WaitGroup

var configParameters configJson
var rows [][]string

func main() {
	log.Println("Start!")

	log.Println("Loading config...")
	bootstrap()

	if 0 == configParameters.Workers {
		log.Print("No workers no love.\n")
		os.Exit(0)
	}

	log.Println("Login...")
	if err := login(); nil != err {
		log.Fatal(err)
	}

	log.Println("Reading input file...")
	file, err := os.Open(configParameters.Report)
	if err != nil {
		fmt.Println("Error opening report file:", err)
		os.Exit(1)
	}
	defer file.Close()

	log.Println("Creating output file...")
	csvFile, err := os.Create(configParameters.Results)
	if err != nil {
		log.Fatalf("Error creating results file (err: %s)", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(file)

	log.Println("Generating workers...")
	urls := make(chan []string, configParameters.Workers)
	for i := 0; i < configParameters.Workers; i++ {
		go worker(urls, i)
	}

	rows = append(rows, []string{"Section", "Days", "Endpoint", "Duration"})

	var section string
	for {
		record, err := reader.Read()
		if err == io.EOF || 0 == configParameters.Limit {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if "" != record[0] {
			section = record[0]
		} else {
			record[0] = section
		}

		fullUrl, err := buildFullUrl(record[1])
		if nil != err || evaluateUrlSkip(fullUrl) || "Section" == section {
			continue
		}

		if "Endpoint" != record[1] && "" != record[1] {
			configParameters.Limit--
			urls <- []string{record[0], fullUrl}
		}
	}

	close(urls)
	wg.Wait()

	w := csv.NewWriter(csvFile)
	w.WriteAll(rows)

	if 0 == configParameters.Limit {
		log.Println("Limit reached.")
	} else {
		log.Println("End.")
	}
}

func bootstrap() {
	configFile, err := ioutil.ReadFile(jsonFileName)
	if nil != err {
		log.Fatalf("Error opening config file: %v", err)
	}

	err = json.Unmarshal(configFile, &configParameters)
	if nil != err {
		log.Fatalf("Error at the unmarshal: %v", err)
	}
}

func login() error {
	return NewClient(configParameters)
}

func buildFullUrl(pathUrl string) (string, error) {
	if "Endpoint" == pathUrl {
		return "", errors.New("Header")
	}

	u := url.URL{}

	u.Scheme = configParameters.Url.Schema
	u.Host = configParameters.Url.Domain
	u.Path = pathUrl

	fullUrl, err := url.PathUnescape(u.String())
	if nil != err {
		return "", err
	}

	return fullUrl, nil
}

func evaluateUrlSkip(u string) bool {
	skip := "" == u

	if !skip && !configParameters.skipCore() {
		parsed, err := url.Parse(u)
		if err != nil {
			log.Fatalf("Error parsing URL: %s (err: %s)", u, err)
		}

		skip = !strings.Contains(parsed.Host, coreDomain)
	}

	return skip
}

func replaceDateTokens(url string, days int) string {
	var parsedUrl string

	now := time.Now()

	parsedUrl = url
	parsedUrl = strings.Replace(parsedUrl,
		configParameters.Tokens.DateFrom,
		now.AddDate(0, 0, days*-1).Format("2006-01-02"), -1)
	parsedUrl = strings.Replace(parsedUrl,
		configParameters.Tokens.DateTo,
		now.Format("2006-01-02"), -1)

	return parsedUrl
}

func hasTokens(url string) bool {
	return strings.Contains(url, "~@")
}

func runUrl(url string) (time.Duration, error) {
	return rc.GetElapsedTime(url)
}

func worker(url <-chan []string, worker int) {
	wg.Add(1)
	defer wg.Done()

	var err error
	var parsedUrl string
	var dayRanges []int
	var elapsedTime time.Duration

	for u := range url {
		if hasTokens(u[1]) {
			dayRanges = configParameters.Days
		} else {
			if configParameters.skipNoDays() {
				continue
			}

			dayRanges = []int{0}
		}

		for _, days := range dayRanges {
			parsedUrl = replaceDateTokens(u[1], days)

			log.Printf("Running %s at worker \"%d\"...\n", parsedUrl, worker+1)
			elapsedTime, err = runUrl(parsedUrl)
			if nil != err {
				fmt.Printf("URL: %s, Error: %s\n", u, err)
				continue
			}

			log.Printf("Worker \"%d\" done %s, elapsed time: %s.\n", worker+1,
				parsedUrl, elapsedTime)

			rows = append(rows, []string{
				u[0],
				fmt.Sprintf("%d days", days),
				parsedUrl,
				fmt.Sprintf("%d", elapsedTime/time.Millisecond)})
		}
	}
}
