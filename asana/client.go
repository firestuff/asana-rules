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
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Section struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Task struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type User struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserTaskList struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Workspace struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type projectsResponse struct {
	Data []*Project `json:"data"`
}

type sectionsResponse struct {
	Data []*Section `json:"data"`
}

type tasksResponse struct {
	Data []*Task `json:"data"`
}

type userResponse struct {
	Data *User `json:"data"`
}

type userTaskListResponse struct {
	Data *UserTaskList `json:"data"`
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

func (c *Client) GetMe() (*User, error) {
	resp := &userResponse{}
	err := c.get("users/me", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetProjects(workspaceGID string) ([]*Project, error) {
	path := fmt.Sprintf("workspaces/%s/projects", workspaceGID)
	resp := &projectsResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetSections(projectGID string) ([]*Section, error) {
	path := fmt.Sprintf("projects/%s/sections", projectGID)
	resp := &sectionsResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTasksFromSection(sectionGID string) ([]*Task, error) {
	path := fmt.Sprintf("sections/%s/tasks", sectionGID)
	resp := &tasksResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetUserTaskList(userGID, workspaceGID string) (*UserTaskList, error) {
	path := fmt.Sprintf("users/%s/user_task_list", userGID)
	values := &url.Values{}
	values.Add("workspace", workspaceGID)
	resp := &userTaskListResponse{}
	err := c.get(path, values, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetWorkspaces() ([]*Workspace, error) {
	resp := &workspacesResponse{}
	err := c.get("workspaces", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Returns one workspace if there is only one
func (c *Client) GetWorkspace() (*Workspace, error) {
	workspaces, err := c.GetWorkspaces()
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

func (s *Section) String() string {
	return fmt.Sprintf("%s (%s)", s.GID, s.Name)
}

func (t *Task) String() string {
	return fmt.Sprintf("%s (%s)", t.GID, t.Name)
}

func (u *User) String() string {
	return fmt.Sprintf("%s (%s <%s>)", u.GID, u.Name, u.Email)
}

func (utl *UserTaskList) String() string {
	return fmt.Sprintf("%s (%s)", utl.GID, utl.Name)
}

func (wrk *Workspace) String() string {
	return fmt.Sprintf("%s (%s)", wrk.GID, wrk.Name)
}
