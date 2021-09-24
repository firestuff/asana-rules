package client

import "fmt"
import "net/url"

type Tag struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type tagsResponse struct {
	Data     []*Tag    `json:"data"`
	NextPage *nextPage `json:"next_page"`
}

func (wc *WorkspaceClient) GetTags() ([]*Tag, error) {
	ret := []*Tag{}

	path := fmt.Sprintf("workspaces/%s/tags", wc.workspace.GID)
	values := &url.Values{}

	for {
		resp := &tagsResponse{}
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
