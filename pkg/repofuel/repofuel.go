package repofuel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	userAgent = "go-client"
)

type Client struct {
	client    *http.Client
	Ingest    *IngestService
	AI        *AIService
	Accounts  *AccountsService
	UserAgent string
}

type service struct {
	client  *Client
	BaseURL *url.URL
}

type Options struct {
	BaseURLs URLSet `yaml:"base_urls"`
}

// TODO: move it
type URLSet map[string]*url.URL

// UnmarshalYAML check at the first if the value existed in the
// environment variables, if omitted, then it check the value in the
// YAML.
func (urls *URLSet) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var raw map[string]string

	if err = unmarshal(&raw); err != nil {
		return err
	}

	if *urls == nil {
		*urls = make(map[string]*url.URL, len(raw))
	}

	for k := range raw {
		v, ok := os.LookupEnv("URL_" + strings.ToUpper(k))
		if !ok {
			v = raw[k]
		}

		(*urls)[k], err = url.Parse(v)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewClient(client *http.Client, opts *Options) *Client {

	c := &Client{
		client:    client,
		UserAgent: "",
	}

	c.Ingest = &IngestService{c, opts.BaseURLs["ingest"]}
	c.AI = &AIService{c, opts.BaseURLs["ai"]}
	c.Accounts = &AccountsService{c, opts.BaseURLs["accounts"]}

	return c
}

func (s *service) NewRequestWithContext(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := s.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return s.client.NewRequestWithContext(ctx, method, u.String(), body)
}

func (s *Client) NewRequestWithContext(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}
	return req, nil
}

type ErrorResponse struct {
	Response *http.Response
	Message  string `json:"message"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			} else if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

func (c *Client) Pip(req *http.Request) (io.ReadCloser, *http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	err = CheckResponse(resp)
	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp, err
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}
