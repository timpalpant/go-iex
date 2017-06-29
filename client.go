package iex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-querystring/query"
)

const baseEndpoint = "https://api.iextrading.com/1.0"

// Client provides methods to interact with IEX's HTTP API for developers.
type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{client}
}

func (c *Client) GetSymbols() ([]*Symbol, error) {
	var result []*Symbol
	err := c.getJSON("/ref-data/symbols", &result)
	return result, err
}

func (c *Client) getJSON(route string, response interface{}) error {
	url := c.endpoint(route)
	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%v: %v", resp.Status, string(body))
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(response)
}

func (c *Client) endpoint(route string) string {
	return baseEndpoint + route
}
