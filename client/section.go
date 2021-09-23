package client

import "fmt"

type Section struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type sectionsResponse struct {
	Data []*Section `json:"data"`
}

type sectionAddTaskData struct {
	Task string `json:"task"`
}

type sectionAddTaskRequest struct {
	Data *sectionAddTaskData `json:"data"`
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

func (wc *WorkspaceClient) AddTaskToSection(task *Task, section *Section) error {
	req := &sectionAddTaskRequest{
		Data: &sectionAddTaskData{
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

func (wc *WorkspaceClient) GetTasksFromSection(section *Section) ([]*Task, error) {
	path := fmt.Sprintf("sections/%s/tasks", section.GID)
	resp := &tasksResponse{}
	err := wc.client.get(path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *Section) String() string {
	return fmt.Sprintf("%s (%s)", s.GID, s.Name)
}
