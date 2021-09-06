package asana

import "encoding/json"
import "io/ioutil"
import "fmt"
import "net/http"
import "net/url"
import "os"

import "github.com/firestuff/asana-rules/headers"

type Client struct {
  client *http.Client
}

type Project struct {
  GID string `json:"gid"`
  Name string `json:"name"`
}

type Task struct {
  GID string `json:"gid"`
  Name string `json:"name"`
}

type User struct {
  GID string `json:"gid"`
  Name string `json:"name"`
  Email string `json:"email"`
}

type Workspace struct {
  GID string `json:"gid"`
  Name string `json:"name"`
}

type projectsResponse struct {
  Data []*Project `json:"data"`
}

type userResponse struct {
  Data *User `json:"data"`
}

type workspacesResponse struct {
  Data []*Workspace `json:"data"`
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

func (c *Client) Me() (*User, error) {
  resp := &userResponse{}
  err := c.get("users/me", nil, resp)
  if err != nil {
    return nil, err
  }
  return resp.Data, nil
}

func (c *Client) Projects(workspaceGID string) ([]*Project, error) {
  resp := &projectsResponse{}
  path := fmt.Sprintf("workspaces/%s/projects", workspaceGID)
  err := c.get(path, nil, resp)
  if err != nil {
    return nil, err
  }
  return resp.Data, nil
}

func (c *Client) Workspaces() ([]*Workspace, error) {
  resp := &workspacesResponse{}
  err := c.get("workspaces", nil, resp)
  if err != nil {
    return nil, err
  }
  return resp.Data, nil
}

// Returns one workspace if there is only one
func (c *Client) Workspace() (*Workspace, error) {
  workspaces, err := c.Workspaces()
  if err != nil {
    return nil, err
  }

  if len(workspaces) != 1 {
    return nil, fmt.Errorf("%d workspaces found", len(workspaces))
  }

  return workspaces[0], nil
}

const baseURL = "https://app.asana.com/api/1.0/"

func (c *Client) get(path string, values *url.Values, out interface{}) error {
  if values == nil {
    values = &url.Values{}
  }
  values.Add("limit", "100")

  url := fmt.Sprintf("%s%s?%s", baseURL, path, values.Encode())
  fmt.Printf("%s\n", url)

  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return err
  }

  resp, err := c.client.Do(req)
  if err != nil {
    return err
  }

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
