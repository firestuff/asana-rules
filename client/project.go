package client

import "fmt"
import "net/url"

type Project struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type projectResponse struct {
	Data *Project `json:"data"`
}

type projectsResponse struct {
	Data     []*Project `json:"data"`
	NextPage *nextPage  `json:"next_page"`
}

func (wc *WorkspaceClient) GetProjects() ([]*Project, error) {
	ret := []*Project{}

	path := fmt.Sprintf("workspaces/%s/projects", wc.workspace.GID)
	values := &url.Values{}

	for {
		resp := &projectsResponse{}
		err := wc.client.get(path, values, resp)
		if err != nil {
			return nil, err
		}

		ret = append(ret, resp.Data...)

		if resp.NextPage == nil {
			break
		}

		values.Set("offset", resp.NextPage.Offset)
	}

	return ret, nil
}

func (p *Project) String() string {
	return fmt.Sprintf("%s (%s)", p.GID, p.Name)
}
