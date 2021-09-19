package asanaclient

import "bytes"
import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "net/url"
import "os"

import "github.com/firestuff/asana-rules/headers"

type Client struct {
	client *http.Client
}

type workspace struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type workspacesResponse struct {
	Data []*workspace `json:"data"`
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

func (c *Client) InWorkspace(name string) (*WorkspaceClient, error) {
	wrk, err := c.getWorkspaceByName(name)
	if err != nil {
		return nil, err
	}

	return &WorkspaceClient{
		client:    c,
		workspace: wrk,
	}, nil
}

func (c *Client) getWorkspaces() ([]*workspace, error) {
	resp := &workspacesResponse{}
	err := c.get("workspaces", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) getWorkspaceByName(name string) (*workspace, error) {
	wrks, err := c.getWorkspaces()
	if err != nil {
		return nil, err
	}

	for _, wrk := range wrks {
		if wrk.Name == name {
			return wrk, nil
		}
	}

	return nil, fmt.Errorf("Workspace `%s` not found", name)
}

const baseURL = "https://app.asana.com/api/1.0/"

func (c *Client) get(path string, values *url.Values, out interface{}) error {
	if values == nil {
		values = &url.Values{}
	}
	values.Add("limit", "100")

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

func (wrk *workspace) String() string {
	return fmt.Sprintf("%s (%s)", wrk.GID, wrk.Name)
}
