package http

import (
	"net/http"
)

var client http.Client

// SetHTTPClient calls the build function on the passed in client builder and sets teh client variable to the return of the builder.
func SetHTTPClient(builder ClientBuilder) {
	client = builder.Build()
}

// GetHTTPClient gets a pointer to the http client.
func GetHTTPClient() *http.Client {
	return &client
}
