package utils

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"

	"github.com/ameenmaali/whoareyou/pkg/config"
)

type Response struct {
	StatusCode    int
	Body          []byte
	Headers       http.Header
	ContentLength int
	GoQueryDoc    *goquery.Document
}

func CreateClient(timeout int) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,
			KeepAlive: time.Second,
		}).DialContext,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout+3) * time.Second,
	}
	return httpClient
}

func SendRequest(u string, config *config.Config) (Response, error) {
	response := Response{}

	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return response, err
	}

	request.Header.Add("User-Agent", browser.Random())

	// Add headers passed in as arguments
	for header, value := range config.Headers {
		request.Header.Add(header, value)
	}

	// Add cookies passed in as arguments
	request.Header.Add("Cookie", config.Cookies)

	resp, err := config.HttpClient.Do(request)

	if err != nil {
		return response, err
	}

	if resp.Body == nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	// Reset the response body to be read again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err == nil {
		response.GoQueryDoc = doc
	}

	response.Body = body
	response.Headers = resp.Header
	response.StatusCode = resp.StatusCode
	response.ContentLength = int(resp.ContentLength)

	return response, err
}
