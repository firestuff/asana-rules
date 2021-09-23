package client

import "fmt"

type User struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

func (u *User) String() string {
	return fmt.Sprintf("%s (%s <%s>)", u.GID, u.Name, u.Email)
}
