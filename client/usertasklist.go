package client

import "fmt"
import "net/url"

// UserTaskLists are actually Projects

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
