package asana

import "fmt"
import "net/http"
import "os"

type withHeader struct {
  Header http.Header
  rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeader{
		Header: make(http.Header),
    rt: rt,
  }
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

func Fetch() {
  c := &http.Client{}

	rt := WithHeader(c.Transport)
	rt.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("ASANA_TOKEN")))
	c.Transport = rt

  resp, err := c.Get("https://app.asana.com/api/1.0/users/me")
  if err != nil {
    panic(err)
  }

  fmt.Printf("%#v\n", resp)
}
