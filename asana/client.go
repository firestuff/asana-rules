package asana

import "encoding/json"
import "io/ioutil"
import "fmt"
import "net/http"
import "net/url"
import "os"
import "strings"

import "cloud.google.com/go/civil"
import "github.com/firestuff/asana-rules/headers"
import "golang.org/x/net/html"

var _TRUE = true
var TRUE = &_TRUE
var _FALSE = false
var FALSE = &_FALSE

type Client struct {
	client *http.Client
}

type SearchQuery struct {
	SectionsAny []*Section
	Completed   *bool
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
	GID             string      `json:"gid"`
	Name            string      `json:"name"`
	DueOn           string      `json:"due_on"`
	ParsedDueOn     *civil.Date `json:"-"`
	HTMLNotes       string      `json:"html_notes"`
	ParsedHTMLNotes *html.Node  `json:"-"`
}

type User struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Workspace struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type projectResponse struct {
	Data *Project `json:"data"`
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

func (c *Client) GetProjects(workspace *Workspace) ([]*Project, error) {
	path := fmt.Sprintf("workspaces/%s/projects", workspace.GID)
	resp := &projectsResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetSections(project *Project) ([]*Section, error) {
	path := fmt.Sprintf("projects/%s/sections", project.GID)
	resp := &sectionsResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTasksFromSection(section *Section) ([]*Task, error) {
	path := fmt.Sprintf("sections/%s/tasks", section.GID)
	resp := &tasksResponse{}
	err := c.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetUserTaskList(user *User, workspace *Workspace) (*Project, error) {
	path := fmt.Sprintf("users/%s/user_task_list", user.GID)
	values := &url.Values{}
	values.Add("workspace", workspace.GID)
	resp := &projectResponse{}
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

func (c *Client) Search(workspace *Workspace, q *SearchQuery) ([]*Task, error) {
	path := fmt.Sprintf("workspaces/%s/tasks/search", workspace.GID)

	values := &url.Values{}

	values.Add("opt_fields", "due_on,html_notes,name")

	if len(q.SectionsAny) > 0 {
		gids := []string{}
		for _, sec := range q.SectionsAny {
			gids = append(gids, sec.GID)
		}
		values.Add("sections.any", strings.Join(gids, ","))
	}

	if q.Completed != nil {
		values.Add("completed", fmt.Sprintf("%t", *q.Completed))
	}

	resp := &tasksResponse{}
	err := c.get(path, values, resp)
	if err != nil {
		return nil, err
	}

	for _, task := range resp.Data {
		err := task.parse()
		if err != nil {
			return nil, err
		}
	}

	return resp.Data, nil

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

func (p *Project) String() string {
	return fmt.Sprintf("%s (%s)", p.GID, p.Name)
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

func (wrk *Workspace) String() string {
	return fmt.Sprintf("%s (%s)", wrk.GID, wrk.Name)
}

func (t *Task) parse() error {
	r := strings.NewReader(t.HTMLNotes)
	root, err := html.Parse(r)
	if err != nil {
		return err
	}
	t.ParsedHTMLNotes = root

	if t.DueOn != "" {
		d, err := civil.ParseDate(t.DueOn)
		if err != nil {
			return err
		}
		t.ParsedDueOn = &d
	}

	return nil
}
