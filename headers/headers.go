package headers

import "net/http"

type Headers struct {
	header http.Header
	rt     http.RoundTripper
}

func NewHeaders(c *http.Client) Headers {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}

	ret := Headers{
		header: http.Header{},
		rt:     c.Transport,
	}

	c.Transport = ret

	return ret
}

func (h *Headers) Add(key, value string) {
	h.header.Add(key, value)
}

func (h Headers) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, vals := range h.header {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}

	return h.rt.RoundTrip(req)
}
