package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"encoding/csv"
	"os"

	"io"

	"io/ioutil"

	"encoding/json"

	"strings"

	"golang.org/x/net/publicsuffix"
)

var wg sync.WaitGroup
var client http.Client
var configParameters configJson
var rows [][]string

func main() {
	log.Println("Start!")

	log.Println("Loading config...")
	bootstrap()

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
	for i := 0; i <= configParameters.Workers; i++ {
		go worker(urls)
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

		if evaluateUrlSkip(record[1]) || "Section" == section {
			continue
		}

		if "Endpoint" != record[1] && "" != record[1] {
			configParameters.Limit--
			urls <- record
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
	var err error

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	jar, err := cookiejar.New(&options)
	if nil == err {
		client = http.Client{Jar: jar}
		_, err = client.PostForm(configParameters.Url.Login.Path, url.Values{
			configParameters.getUsernameField(): {
				configParameters.getUsernameValue()},
			configParameters.getPasswordField(): {
				configParameters.getPasswordValue()},
		})
	}

	return err
}

func evaluateUrlSkip(u string) bool {
	skip := "" == u

	if !skip && !configParameters.skipCore() {
		parsed, err := url.Parse(u)
		if err != nil {
			log.Fatalf("Error parsing URL: %s (err: %s)", u, err)
		}

		skip = !strings.Contains(parsed.Host, configParameters.Domain)
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
	var err error

	start := time.Now()
	_, err = client.Get(url)
	if nil != err {
		return time.Since(start), err
	}

	return time.Since(start), nil
}

func worker(url <-chan []string) {
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
			if configParameters.SkipNoDays {
				continue
			}

			dayRanges = []int{0}
		}

		for _, days := range dayRanges {
			parsedUrl = replaceDateTokens(u[1], days)

			log.Printf("Running %s...\n", parsedUrl)
			elapsedTime, err = runUrl(parsedUrl)
			if nil != err {
				fmt.Printf("URL: %s, Error: %s\n", u, err)
				continue
			}

			rows = append(rows, []string{
				u[0],
				fmt.Sprintf("%d days", days),
				parsedUrl,
				fmt.Sprintf("%d", elapsedTime/time.Millisecond)})
		}
	}
}
