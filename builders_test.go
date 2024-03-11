package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ClientBuilder_check_defaults(t *testing.T) {
	c := NewClientBuilder().Build()

	assert.Equal(t, time.Duration(100)*time.Second, c.Timeout)
}

func Test_RequestBuilder(t *testing.T) {
	req, err := NewRequestBuilder().
		Method(http.MethodGet).
		Url("test.com").
		AddAuthorization("authValue").
		AddHeader("header1", "headerValue").
		AddParam("param1", "value1").
		AddParam("param2", "value2").
		Build()

	assert.Equal(t, nil, err)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "authValue", req.Header.Get("Authorization"))
	assert.Equal(t, "headerValue", req.Header.Get("header1"))
	assert.Equal(t, "value1", req.URL.Query().Get("param1"))
	assert.Equal(t, "value2", req.URL.Query().Get("param2"))
}
