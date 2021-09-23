package client

import "bytes"
import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "net/url"
import "os"

import "github.com/firestuff/automana/headers"

type Client struct {
	client *http.Client
}

type errorDetails struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Errors []*errorDetails `json:"errors"`
}

type emptyResponse struct {
	Data interface{} `json:"data"`
}

func NewClient(token string) *Client {
	c := &Client{
		client: &http.Client{},
	}

	hdrs := headers.NewHeaders(c.client)
	hdrs.Add("Accept", "application/json")
	hdrs.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return c
}

func NewClientFromEnv() *Client {
	return NewClient(os.Getenv("ASANA_TOKEN"))
}

const baseURL = "https://app.asana.com/api/1.0/"
const perPage = 100

func (c *Client) get(path string, values *url.Values, out interface{}) error {
	if values == nil {
		values = &url.Values{}
	}
	values.Set("limit", fmt.Sprintf("%d", perPage))

	url := fmt.Sprintf("%s%s?%s", baseURL, path, values.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	if resp.StatusCode != 200 {
		errorResp := &errorResponse{}
		err = dec.Decode(errorResp)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", resp.Status, errorResp.Errors[0].Message)
	}

	err = dec.Decode(out)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) post(path string, body interface{}, out interface{}) error {
	return c.doWithBody("POST", path, body, out)
}

func (c *Client) put(path string, body interface{}, out interface{}) error {
	return c.doWithBody("PUT", path, body, out)
}

func (c *Client) doWithBody(method string, path string, body interface{}, out interface{}) error {
	url := fmt.Sprintf("%s%s", baseURL, path)

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	err := enc.Encode(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(out)
	if err != nil {
		return err
	}

	return nil
}
