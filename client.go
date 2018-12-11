package arxiv

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

// A Client communicates with the arXiv API
type Client struct {
	Eprints    EprintsService
	BaseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new HTTP Client for arXiv.  If httpClient == nil,
// then http.DefaultClient is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{
		BaseURL: &url.URL{
			Scheme: "http",
			Host:   "export.arxiv.org",
			Path:   "/api/",
		},
		httpClient: httpClient,
	}
	c.Eprints = &eprintsService{c}

	return c
}

// QueryOptions specifies general pagination options for fetching a list of
// results.
type QueryOptions struct {
	MaxResults int    `url:"max_results,omitempty"`
	Start      int    `url:"start,omitempty"`
	SortBy     string `url:"sortBy,omitempty"`
	SortOrder  string `url:"sortOrder,omitempty"`
}

// DefaultMaxResults is the default number of items to return for a query
const DefaultMaxResults = 100

// url generates the URL to the named thesrc API endpoint, using the
// specified route variables and query options.
func (c *Client) url(apiRouteName string, opt interface{}) (*url.URL, error) {
	url := &url.URL{
		Path: apiRouteName,
	}

	if opt != nil {
		err := addOptions(url, opt)
		if err != nil {
			return nil, err
		}
	}

	return url, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client. Relative
// URLs should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// XML-decoded and stored in the value pointed to by v, or returned as an error
// if an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if v != nil {
		if bp, ok := v.(*[]byte); ok {
			*bp, err = ioutil.ReadAll(resp.Body)
		} else {
			err = xml.NewDecoder(resp.Body).Decode(v)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
	}
	return resp, nil
}

// addOptions adds the parameters in opt as URL query parameters to u. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(u *url.URL, opt interface{}) error {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	qs, err := query.Values(opt)
	if err != nil {
		return err
	}

	u.RawQuery = qs.Encode()
	return nil
}
