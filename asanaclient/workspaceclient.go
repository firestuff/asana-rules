package asanaclient

import "fmt"
import "net/url"
import "strings"

import "cloud.google.com/go/civil"
import "golang.org/x/net/html"

var _TRUE = true
var TRUE = &_TRUE
var _FALSE = false
var FALSE = &_FALSE

type WorkspaceClient struct {
	client    *Client
	workspace *workspace
}

type SearchQuery struct {
	SectionsAny []*Section
	Completed   *bool
	DueOn       *civil.Date
	DueBefore   *civil.Date
	DueAfter    *civil.Date
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

type addTaskDetails struct {
	Task string `json:"task"`
}

type addTaskRequest struct {
	Data *addTaskDetails `json:"data"`
}

type emptyResponse struct {
	Data interface{} `json:"data"`
}

type errorDetails struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Errors []*errorDetails `json:"errors"`
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

func (wc *WorkspaceClient) GetMe() (*User, error) {
	resp := &userResponse{}
	err := wc.client.get("users/me", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) AddTaskToSection(task *Task, section *Section) error {
	req := &addTaskRequest{
		Data: &addTaskDetails{
			Task: task.GID,
		},
	}

	resp := &emptyResponse{}

	path := fmt.Sprintf("sections/%s/addTask", section.GID)
	err := wc.client.post(path, req, resp)
	if err != nil {
		return err
	}

	return nil
}

func (wc *WorkspaceClient) GetProjects() ([]*Project, error) {
	path := fmt.Sprintf("workspaces/%s/projects", wc.workspace.GID)
	resp := &projectsResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) GetSections(project *Project) ([]*Section, error) {
	path := fmt.Sprintf("projects/%s/sections", project.GID)
	resp := &sectionsResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) GetSectionsByName(project *Project) (map[string]*Section, error) {
	secs, err := wc.GetSections(project)
	if err != nil {
		return nil, err
	}

	secsByName := map[string]*Section{}
	for _, sec := range secs {
		secsByName[sec.Name] = sec
	}

	return secsByName, err
}

func (wc *WorkspaceClient) GetSectionByName(project *Project, name string) (*Section, error) {
	secsByName, err := wc.GetSectionsByName(project)
	if err != nil {
		return nil, err
	}

	sec, found := secsByName[name]
	if !found {
		return nil, fmt.Errorf("Section '%s' not found", name)
	}

	return sec, nil
}

func (wc *WorkspaceClient) GetTasksFromSection(section *Section) ([]*Task, error) {
	path := fmt.Sprintf("sections/%s/tasks", section.GID)
	resp := &tasksResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) GetUserTaskList(user *User) (*Project, error) {
	path := fmt.Sprintf("users/%s/user_task_list", user.GID)
	values := &url.Values{}
	values.Add("workspace", wc.workspace.GID)
	resp := &projectResponse{}
	err := wc.client.get(path, values, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) GetMyUserTaskList() (*Project, error) {
	me, err := wc.GetMe()
	if err != nil {
		return nil, err
	}

	return wc.GetUserTaskList(me)
}

func (wc *WorkspaceClient) Search(q *SearchQuery) ([]*Task, error) {
	path := fmt.Sprintf("workspaces/%s/tasks/search", wc.workspace.GID)

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

	if q.DueOn != nil {
		values.Add("due_on", q.DueOn.String())
	}

	if q.DueBefore != nil {
		values.Add("due_on.before", q.DueBefore.String())
	}

	if q.DueAfter != nil {
		values.Add("due_on.after", q.DueAfter.String())
	}

	resp := &tasksResponse{}
	err := wc.client.get(path, values, resp)
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
