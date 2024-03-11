package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_set_and_get_http_client(t *testing.T) {
	SetHTTPClient(NewClientBuilder().Timeout(100))
	assert.Equal(t, time.Duration(100)*time.Second, GetHTTPClient().Timeout)
}
