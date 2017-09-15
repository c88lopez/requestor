package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"io/ioutil"

	"log"

	"golang.org/x/net/publicsuffix"
)

var client http.Client
var rc RealClient

type RealClient struct {
	Client http.Client

	Headers [][]string
}

type WebClient interface {
	Get(string) (d time.Duration, err error)
}

type Client struct {
	Http WebClient
}

func NewClient(cp configJson) error {
	var err error

	if cp.Login.Enabled {
		options := cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		}

		jar, err := cookiejar.New(&options)
		if nil == err {
			loginUrl, err := buildFullUrl(cp.Login.Path)

			if nil == err {
				client = http.Client{Jar: jar}
				_, err = client.PostForm(loginUrl, url.Values{
					cp.getUsernameField(): {
						cp.getUsernameValue()},
					cp.getPasswordField(): {
						cp.getPasswordValue()},
				})
			}
		}
	} else {
		client = http.Client{}
	}

	rc.Client = client

	rc.Headers = cp.Headers

	return err
}

func (c *RealClient) GetElapsedTime(url string) (time.Duration, error) {
	start := time.Now()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return time.Since(start), err
	}

	for _, value := range c.Headers {
		request.Header.Set(value[0], value[1])
	}

	response, err := rc.Client.Do(request)
	if err != nil {
		return time.Since(start), err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return time.Since(start), err
	}

	log.Printf("Response: %s\n", responseData)

	return time.Since(start), err
}
