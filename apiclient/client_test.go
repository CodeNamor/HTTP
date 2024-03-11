package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
)

const (
	baseURLSuffix = "v2"
)

const testURL string = `https://test.com/`

func newClient() RetryClient {

	c := http.Client{}

	rc := NewExtendedHTTPClient(1, &c)

	return rc
}

func TestApiClient_InitClient(t *testing.T) {
	c, _ := InitClient(newClient(), testURL, "test", false, "")
	if c.UserAgent != "test" {
		t.Errorf("Init client wrong user agent, expecting %v, got %v", "test", c.UserAgent)
	}
}

func TestApiClient_Do_httpInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rc := NewMockRetryClient(ctrl)

	rc.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
	}, nil)

	c, _ := InitClient(rc, testURL, "test", false, "")
	request, err := http.NewRequest(http.MethodGet, "http://test", nil)
	if err != nil {
		t.Errorf("Do unable to create request")
		return
	}
	request.URL = nil
	resp, err := c.Do(context.Background(), request)
	if resp == nil {
		t.Errorf("Do expected response to not be nil")
		return
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Do expected http internal server error")
	}
}

func TestApiClient_Get_VerifyRequestURL(t *testing.T) {
	executeRequest := func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error) {
		resp, err := client.Get(ctx, path, queryParams)
		// intentionally return response and error objects As-Is
		return resp, err
	}

	v := url.Values{}
	v.Set("name", "Ava")
	v.Set("friend", "Jess")

	verifyRequestURL(t, "", nil, executeRequest)
	verifyRequestURL(t, "", &v, executeRequest)
	verifyRequestURL(t, baseURLSuffix, nil, executeRequest)
	verifyRequestURL(t, baseURLSuffix, &v, executeRequest)
}

func TestApiClient_Post_VerifyRequestURL(t *testing.T) {

	executeRequest := func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error) {
		resp, err := client.Post(ctx, path, body)
		// intentionally return response and error objects As-Is
		return resp, err
	}

	verifyRequestURL(t, "", nil, executeRequest)
	verifyRequestURL(t, baseURLSuffix, nil, executeRequest)
}

func TestApiClient_PostWithQueryParams_VerifyRequestURL(t *testing.T) {

	executeRequest := func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error) {
		resp, err := client.PostWithQueryParams(ctx, path, queryParams, body)
		// intentionally return response and error objects As-Is
		return resp, err
	}

	v := url.Values{}
	v.Set("name", "Ava")
	v.Set("friend", "Jess")

	verifyRequestURL(t, "", nil, executeRequest)
	verifyRequestURL(t, baseURLSuffix, &v, executeRequest)
}

func TestApiClient_Put_VerifyRequestURL(t *testing.T) {

	executeRequest := func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error) {
		resp, err := client.Put(ctx, path, body)
		// intentionally return response and error objects As-Is
		return resp, err
	}

	verifyRequestURL(t, "", nil, executeRequest)
	verifyRequestURL(t, baseURLSuffix, nil, executeRequest)
}

func TestApiClient_Delete_VerifyRequestURL(t *testing.T) {

	executeRequest := func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error) {
		resp, err := client.Delete(ctx, path, body)
		// intentionally return response and error objects As-Is
		return resp, err
	}

	verifyRequestURL(t, "", nil, executeRequest)
	verifyRequestURL(t, baseURLSuffix, nil, executeRequest)
}

type makeRequest func(ctx context.Context, client APIClient, path string, queryParams *url.Values, body io.Reader) (*Response, error)

func verifyRequestURL(t *testing.T, baseURLSuffix string, queryParams *url.Values, request makeRequest) {
	URLSuffix := "roles/bus"
	var uri string
	if baseURLSuffix == "" {
		uri = "/" + URLSuffix
	} else {
		uri = path.Join("/", baseURLSuffix, URLSuffix)
	}
	if queryParams != nil {
		uri += "?" + queryParams.Encode()
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case uri:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		default:
			t.Errorf("Wanted URL path: %v, got instead: %v", uri, r.RequestURI)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	ctx, c, teardown := getServer(t, h, baseURLSuffix)
	defer teardown()

	roles := []string{"LOL"}
	r := struct {
		Roles []string `json:"roles"`
	}{
		roles,
	}
	body, _ := json.Marshal(&r)

	resp, err := request(ctx, c, URLSuffix, queryParams, bytes.NewReader(body))
	if err != nil {
		t.Errorf("Expected no error, got error instead: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status code %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func getServer(t *testing.T, handlerFunc http.HandlerFunc, baseURLPrefix string) (context.Context, APIClient, func()) {
	s := httptest.NewServer(handlerFunc)

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatalf("Invalid test server url: %v", err)
	}
	u.Path = path.Join(u.Path, baseURLPrefix)
	s.URL = u.String()

	c, err := InitClient(newClient(), u.String(), s.URL, false, "")
	ctx := context.Background()
	return ctx, c, func() {
		s.Close()
	}
}
