package client

import "fmt"

type Project struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type projectResponse struct {
	Data *Project `json:"data"`
}

type projectsResponse struct {
	Data []*Project `json:"data"`
}

func (wc *WorkspaceClient) GetProjects() ([]*Project, error) {
	// TODO: Handle pagination
	path := fmt.Sprintf("workspaces/%s/projects", wc.workspace.GID)
	resp := &projectsResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (p *Project) String() string {
	return fmt.Sprintf("%s (%s)", p.GID, p.Name)
}
