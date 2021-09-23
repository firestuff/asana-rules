package client

import "fmt"

type Workspace struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type workspacesResponse struct {
	Data []*Workspace `json:"data"`
}

func (c *Client) InWorkspace(name string) (*WorkspaceClient, error) {
	wrk, err := c.GetWorkspaceByName(name)
	if err != nil {
		return nil, err
	}

	return &WorkspaceClient{
		client:    c,
		workspace: wrk,
	}, nil
}

func (c *Client) GetWorkspaces() ([]*Workspace, error) {
	// TODO: Handle pagination
	resp := &workspacesResponse{}
	err := c.get("workspaces", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetWorkspaceByName(name string) (*Workspace, error) {
	wrks, err := c.GetWorkspaces()
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

func (wrk *Workspace) String() string {
	return fmt.Sprintf("%s (%s)", wrk.GID, wrk.Name)
}
