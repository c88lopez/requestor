package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

var client http.Client
var rc RealClient

type RealClient struct {
	Client http.Client
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
			client = http.Client{Jar: jar}
			_, err = client.PostForm(cp.Url.Login.Path, url.Values{
				cp.getUsernameField(): {
					cp.getUsernameValue()},
				cp.getPasswordField(): {
					cp.getPasswordValue()},
			})
		}
	} else {
		client = http.Client{}
	}

	rc.Client = client

	return err
}

func (c *RealClient) GetElapsedTime(url string) (time.Duration, error) {
	start := time.Now()

	_, err := rc.Client.Get(url)

	elapsedTime := time.Since(start)

	return elapsedTime, err
}
