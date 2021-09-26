package client

type WorkspaceClient struct {
	client          *Client
	workspace       *Workspace
	rateLimitSearch *RateLimit
}
