package apiclient

import (
	"context"
	"errors"
	cLog "github.com/CodeNamor/custom_logging"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// APIClient base apiClient interface
type APIClient interface {
	Get(ctx context.Context, path string, queryParams *url.Values) (*Response, error)
	Do(ctx context.Context, request *http.Request) (*Response, error)
	Post(ctx context.Context, path string, body io.Reader) (*Response, error)
	PostWithQueryParams(ctx context.Context, urlPath string, queryParams *url.Values, body io.Reader) (*Response, error)
	Put(ctx context.Context, path string, body io.Reader) (*Response, error)
	Delete(ctx context.Context, path string, body io.Reader) (*Response, error)
	PostXML(ctx context.Context, path string, body io.Reader, soapAction string) (*Response, error)
}

// Client holds the baseURL, client and userAgent.
type Client struct {
	BaseURL               *url.URL
	UserAgent             string
	RequiresAuthorization bool
	AuthHeaderName        string
	AuthKey               string
	HTTPClient            RetryClient
}

// Response is the basic response from the APIClient
type Response struct {
	Body            []byte
	StatusCode      int
	OriginalRequest *http.Request
	FaultString     string
}

// InitClient inits the client given the params passed in.
func InitClient(httpClient RetryClient, baseURL string, userAgent string, reqAuth bool, authKey string) (*Client, error) {

	baseEndpoint, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		BaseURL:               baseEndpoint,
		UserAgent:             userAgent,
		RequiresAuthorization: reqAuth,
		AuthKey:               authKey,
		HTTPClient:            httpClient,
	}

	return c, nil
}

func (c *Client) PostXML(ctx context.Context, urlPath string, body io.Reader, soapAction string) (*Response, error) {
	u := *c.BaseURL
	u.Path = path.Join(u.Path, urlPath)
	request, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		log.Errorf("error creating POST XML request: %v", err.Error())

		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	request.Header.Set("Content-Type", "text/xml;charset=utf-8")
	request.Header.Set("SOAPAction", soapAction)
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}

// Get basic HTTP get call with support for request parameters and query parameters
func (c *Client) Get(ctx context.Context, urlPath string, queryParams *url.Values) (*Response, error) {
	var u url.URL
	if queryParams != nil && len(*queryParams) > 0 {
		q := c.BaseURL.Query()
		for key, value := range *queryParams {
			q[key] = append([]string{}, value...)
		}
		u = *c.BaseURL
		u.Path = path.Join(u.Path, urlPath)
		u.RawQuery = q.Encode()
	} else {
		u = *c.BaseURL
		u.Path = path.Join(u.Path, urlPath)
	}
	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Errorf("error creating GET request: %v", err.Error())
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}

// Do executes a HTTP request
func (c *Client) Do(ctx context.Context, request *http.Request) (*Response, error) {
	if request != nil {
		log.Debugf("APIClient Do(): method %v, url %v", request.Method, request.URL)
	} else {
		log.Errorf("client.DO request was NIL, can not execute do request")
		return nil, errors.New("request was NIL, can not execute do request")
	}
	if c.RequiresAuthorization {
		if c.AuthHeaderName != "" {
			request.Header.Set(c.AuthHeaderName, c.AuthKey)
		} else {
			request.Header.Set("Authorization", c.AuthKey)
		}

	}

	var resp = &Response{}

	request = request.WithContext(ctx)

	var response *http.Response
	request.Close = true
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		log.Errorf("Error sending HTTP request to %s: %v", request.URL, err.Error())
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		default:
		}

		return resp, err
	}

	defer response.Body.Close()
	if response.Body != nil {
		resp.Body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			resp.StatusCode = http.StatusInternalServerError
			log.WithFields(cLog.FieldsFromCTX(ctx)).Errorf("Error reading HTTP response body: %v\n", err)
			return resp, err
		}
	}
	resp.OriginalRequest = request
	resp.StatusCode = response.StatusCode
	return resp, err
}

// Put creates a put request and calls Do
func (c *Client) Put(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	u := *c.BaseURL
	u.Path = path.Join(u.Path, urlPath)
	request, err := http.NewRequest(http.MethodPut, u.String(), body)
	if err != nil {
		log.Errorf("failed to create PUT request: %v", err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}

// Delete creates a Delete request and calls Do
func (c *Client) Delete(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	u := *c.BaseURL
	u.Path = path.Join(u.Path, urlPath)
	request, err := http.NewRequest(http.MethodDelete, u.String(), body)
	if err != nil {
		log.Errorf("failed to create DELETE request: %v", err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}

// Post creates a post request and calls Do
func (c *Client) Post(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	u := *c.BaseURL
	u.Path = path.Join(u.Path, urlPath)
	request, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		log.Errorf("failed to create POST : %v", err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}

// PostWithQueryParams creates a post request and calls Do
func (c *Client) PostWithQueryParams(ctx context.Context, urlPath string, queryParams *url.Values, body io.Reader) (*Response, error) {
	var u url.URL
	if queryParams != nil && len(*queryParams) > 0 {
		q := c.BaseURL.Query()
		for key, value := range *queryParams {
			q[key] = append([]string{}, value...)
		}
		u = *c.BaseURL
		u.Path = path.Join(u.Path, urlPath)
		u.RawQuery = q.Encode()
	} else {
		u = *c.BaseURL
		u.Path = path.Join(u.Path, urlPath)
	}
	request, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		log.Errorf("failed to create POST : %v", err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	return c.Do(ctx, request)
}
