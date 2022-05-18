package client

import "fmt"
import "strings"

import "cloud.google.com/go/civil"
import "golang.org/x/net/html"

type Task struct {
	GID             string           `json:"gid,omitempty"`
	Name            string           `json:"name,omitempty"`
	CreatedAt       string           `json:"created_at,omitempty"`
	DueOn           string           `json:"due_on,omitempty"`
	ParsedDueOn     *civil.Date      `json:"-"`
	HTMLNotes       string           `json:"html_notes,omitempty"`
	ParsedHTMLNotes *html.Node       `json:"-"`
	AssigneeSection *AssigneeSection `json:"assignee_section"`
}

type AssigneeSection struct {
	GID string `json:"gid,omitempty"`
}

type taskResponse struct {
	Data *Task `json:"data"`
}

type tasksResponse struct {
	Data     []*Task   `json:"data"`
	NextPage *nextPage `json:"next_page"`
}

type taskUpdate struct {
	Data *Task `json:"data"`
}

func (wc *WorkspaceClient) UpdateTask(task *Task) error {
	path := fmt.Sprintf("tasks/%s", task.GID)

	task.GID = ""

	update := &taskUpdate{
		Data: task,
	}

	resp := &taskResponse{}
	err := wc.client.put(path, update, resp)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) String() string {
	return fmt.Sprintf("%s (%s)", t.GID, t.Name)
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
