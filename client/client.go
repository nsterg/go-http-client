package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Client is a wrapper for http client performing http requests and handling http responses
type Client struct {
	HTTPClient HTTPClient
	BaseURL    string
}

// HTTPClient is an interface to allow mocking of httpClient's Do method
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SendAndConsume sends an http.Request based using the provided url, http method and payload
// Parses httpResponse body and assigns it to the provided success or error response
func (c *Client) SendAndConsume(url string, method string, payload, success, failure interface{}) (*http.Response, error) {
	req, err := c.createReq(url, method, payload)
	if err != nil {
		log.Printf("Failed to create http request. Error was: %s", err.Error())
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Failed to send http request. Error was: %s", err.Error())
		return resp, err
	}

	defer resp.Body.Close()

	err = c.consume(resp, success, failure)
	return resp, err
}

func (c *Client) consume(resp *http.Response, success, failure interface{}) error {
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if isUnsuccessfulStatusCode(resp) {
		err = json.Unmarshal(respBody, failure)
		if err != nil {
			log.Printf("Failed to umarshal error response. Error was: %s", err.Error())
			return err
		}
		log.Printf("Request failed due to statusCode: %d", resp.StatusCode)
		return nil
	}

	if len(respBody) > 0 {
		err = json.Unmarshal(respBody, success)
		if err != nil {
			log.Printf("Failed to umarshal success response. Error was: %s", err.Error())
			return err
		}
	}
	return nil
}

func (c *Client) createReq(url string, method string, payload interface{}) (*http.Request, error) {
	reqURL := c.BaseURL + url
	var data io.Reader
	if payload != nil {
		jsonReq, _ := json.Marshal(payload)

		data = bytes.NewBuffer(jsonReq)
	}

	req, err := http.NewRequest(method, reqURL, data)
	if err != nil {
		log.Printf("Failed to create http request. Error was: %s", err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func isUnsuccessfulStatusCode(resp *http.Response) bool {
	return !(resp.StatusCode >= 200 && resp.StatusCode < 300)
}
