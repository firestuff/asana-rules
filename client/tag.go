package client

import "fmt"

type Tag struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type tagsResponse struct {
	Data []*Tag `json:"data"`
}

func (wc *WorkspaceClient) GetTags() ([]*Tag, error) {
	// TODO: Handle pagination
	path := fmt.Sprintf("workspaces/%s/tags", wc.workspace.GID)
	resp := &tagsResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (wc *WorkspaceClient) GetTagsByName() (map[string]*Tag, error) {
	tags, err := wc.GetTags()
	if err != nil {
		return nil, err
	}

	tagsByName := map[string]*Tag{}
	for _, tag := range tags {
		tagsByName[tag.Name] = tag
	}

	return tagsByName, err
}
