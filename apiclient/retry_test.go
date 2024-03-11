package apiclient

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_NewExtendedHTTPClient(t *testing.T) {
	testCases := []struct {
		maxRetries int
		statusCode int
		body       string
	}{ // for maxRetries, expected statusCode and body
		{maxRetries: -1, statusCode: 500},
		{maxRetries: 0, statusCode: 500},
		{maxRetries: 1, statusCode: 500},
		{maxRetries: 2, statusCode: 500},
		{maxRetries: 3, statusCode: 200, body: "hello"},
		{maxRetries: 4, statusCode: 200, body: "hello"},
	}
	for _, tc := range testCases {
		subTestName := fmt.Sprintf("maxRetries:%d", tc.maxRetries) // tc info
		t.Run(subTestName, func(t *testing.T) {
			requestCount := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				if requestCount <= 2 { // fails on first two attempts
					w.WriteHeader(500) // unexpected error
					return
				}
				_, i := fmt.Fprint(w, "hello")
				if i != nil {
					return
				}
			}))
			defer ts.Close()
			client := NewExtendedHTTPClient(tc.maxRetries, ts.Client())
			require.NotNil(t, client)
			req, err := http.NewRequest("POST", ts.URL, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tc.statusCode, resp.StatusCode)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tc.body, string(body))
		})
	}

}

func Test_createLogHook(t *testing.T) {
	testCases := []struct {
		maxRetries int
		expected   string
	}{ // for maxRetries, expect logContains1 logContains2, otherwise empty
		{maxRetries: 0},
		{maxRetries: 1},
		{maxRetries: 2, expected: "level=warning msg=\"attempt:1 retrying: myerror\"\n"},
		{maxRetries: 3, expected: "level=warning msg=\"attempt:2 retrying: myerror\"\n"},
	}

	for _, tc := range testCases {
		subTestName := fmt.Sprintf("maxRetries:%d", tc.maxRetries)
		t.Run(subTestName, func(t *testing.T) {
			log.SetLevel(log.WarnLevel)
			log.SetFormatter(&log.TextFormatter{})

			buffer := &bytes.Buffer{}
			log.SetOutput(buffer)
			f := func(buffer *bytes.Buffer) {

			}
			defer f(buffer)
			logHook := createLogHook(tc.maxRetries)
			for i := 0; i < 5; i++ { // simulate error calls to logHook
				logHook(pester.ErrEntry{
					Attempt: i + 1,
					Err:     fmt.Errorf("myerror"),
				})
			}
			require.Contains(t, buffer.String(), tc.expected)
		})
	}
}

func ExampleNewExtendedHTTPClient() {
	const maxRetries = 2 // total attempts
	const insecureSkipVerify = true

	// setup any tls configuration necessary
	tlsConfig := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	// if !insecureSkipVerify {
	// 	tlsConfig.RootCAs = config.CertPool
	// }

	// pull this value from the proper part of config
	const timeoutSecs = 30
	timeout := time.Duration(timeoutSecs * int(time.Second))

	// create an *http.RetryClient that uses tls and timeout configs
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   timeout,
	}

	// create an *http.RetryClient compatible instance from an existing
	// *http.RetryClient that includes retry capabilities
	retryClient := NewExtendedHTTPClient(maxRetries, client)

	// examining the resultant client
	rcValue := reflect.ValueOf(retryClient)
	fmt.Println("retryClient is not nil:", !rcValue.IsNil())
	fmt.Println("retryClient.Do method is not nil:", !rcValue.MethodByName("Do").IsNil())
	// Output: retryClient is not nil: true
	// retryClient.Do method is not nil: true
}
