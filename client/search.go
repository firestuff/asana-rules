package client

import "fmt"
import "net/url"
import "strings"

import "cloud.google.com/go/civil"

type SearchQuery struct {
	SectionsAny []*Section
	Completed   *bool
	Due         *bool
	DueOn       *civil.Date
	DueBefore   *civil.Date
	DueAfter    *civil.Date
	TagsAny     []*Tag
	TagsNot     []*Tag
}

var _TRUE = true
var TRUE = &_TRUE
var _FALSE = false
var FALSE = &_FALSE

func (wc *WorkspaceClient) Search(q *SearchQuery) ([]*Task, error) {
	path := fmt.Sprintf("workspaces/%s/tasks/search", wc.workspace.GID)

	values := &url.Values{
		"sort_by":        []string{"created_at"},
		"sort_ascending": []string{"true"},
	}

	values.Add("opt_fields", "created_at,due_on,html_notes,name")

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

	if q.Due != nil {
		if *q.Due {
			values.Add("due_on.after", "1970-01-01")
		} else {
			values.Add("due_on", "null")
		}
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

	if len(q.TagsAny) > 0 {
		gids := []string{}
		for _, sec := range q.TagsAny {
			gids = append(gids, sec.GID)
		}
		values.Add("tags.any", strings.Join(gids, ","))
	}

	if len(q.TagsNot) > 0 {
		gids := []string{}
		for _, sec := range q.TagsNot {
			gids = append(gids, sec.GID)
		}
		values.Add("tags.not", strings.Join(gids, ","))
	}

	tasksByGID := map[string]*Task{}

	for {
		resp := &tasksResponse{}
		err := wc.client.get(path, values, resp)
		if err != nil {
			return nil, err
		}

		maxCreatedAt := ""

		for _, task := range resp.Data {
			err := task.parse()
			if err != nil {
				return nil, err
			}
			tasksByGID[task.GID] = task

			if task.CreatedAt > maxCreatedAt {
				maxCreatedAt = task.CreatedAt
			}
		}

		if len(resp.Data) < perPage {
			break
		}

		values.Set("created_at.after", maxCreatedAt)
	}

	tasks := []*Task{}
	for _, task := range tasksByGID {
		tasks = append(tasks, task)
	}

	return tasks, nil
}
