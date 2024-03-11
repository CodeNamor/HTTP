package http

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"time"
)

// ClientBuilder is an interface implemented by various functions.
type ClientBuilder interface {
	Timeout(seconds int) ClientBuilder
	IdleConnTimeout(seconds int) ClientBuilder
	MaxIdleConnPerHost(connections int) ClientBuilder
	MaxConnPerHost(connections int) ClientBuilder
	DisableCompression(flag bool) ClientBuilder
	MaxRetries(tries int) ClientBuilder
	InsecureSkipVerify(flag bool) ClientBuilder
	PemCertificates([]byte) ClientBuilder
	Build() http.Client
}

type clientBuilder struct {
	timeout               int
	idleConnTimeout       int
	maxIdleConnPerHost    int
	maxConnPerHost        int
	disableCompression    bool
	maxRetries            int
	tlsInsecureSkipVerify bool
	pemCertificates       []byte
}

// NewClientBuilder constructs a new instance of ClientBuilder with default values.
func NewClientBuilder() ClientBuilder {
	return &clientBuilder{
		timeout:               100,
		idleConnTimeout:       30,
		maxIdleConnPerHost:    16,
		maxConnPerHost:        32,
		disableCompression:    false,
		maxRetries:            2,
		tlsInsecureSkipVerify: false,
		pemCertificates:       []byte{},
	}
}

// Timeout receives an integer that represents seconds and assigns that value to builder's timeout field.
func (b *clientBuilder) Timeout(seconds int) ClientBuilder {
	b.timeout = seconds
	return b
}

// IdleConnTimeout receives an integer that represents seconds and assigns that value to builder's IdleConnTimeout field.
func (b *clientBuilder) IdleConnTimeout(seconds int) ClientBuilder {
	b.idleConnTimeout = seconds
	return b
}

// MaxIdleConnPerHost receives an integer that represents connections and assigns that value to builder's MaxIdleConnPerHost field.
func (b *clientBuilder) MaxIdleConnPerHost(connections int) ClientBuilder {
	b.maxIdleConnPerHost = connections
	return b
}

// MaxConnPerHost receives an integer that represents connections and assigns that value to builder's MaxConnPerHost field.
func (b *clientBuilder) MaxConnPerHost(connections int) ClientBuilder {
	b.maxConnPerHost = connections
	return b
}

// DisableCompression receives a boolean value and assigns that value to builder's DisableCompression field.
func (b *clientBuilder) DisableCompression(flag bool) ClientBuilder {
	b.disableCompression = flag
	return b
}

// MaxRetries receives an integer that represents attempts and assigns that value to builder's MaxRetries field.
func (b *clientBuilder) MaxRetries(tries int) ClientBuilder {
	b.maxRetries = tries
	return b
}

// InsecureSkipVerify receives a boolean value and assigns that value to builder's tlsInsecureSkipVerify field.
func (b *clientBuilder) InsecureSkipVerify(flag bool) ClientBuilder {
	b.tlsInsecureSkipVerify = flag
	return b
}

// PemCertificates receives a byte array that represents PEM Certificates and assigns that array to builder's pemCertificates field.
func (b *clientBuilder) PemCertificates(pemCerts []byte) ClientBuilder {
	b.pemCertificates = pemCerts
	return b
}

// Build creates and returns an instantiated http client.
func (b *clientBuilder) Build() http.Client {
	httpClient := http.Client{
		Timeout: time.Second * time.Duration(b.timeout),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: b.tlsInsecureSkipVerify,
				RootCAs:            getCertPool(b.tlsInsecureSkipVerify, b.pemCertificates),
			},
		},
	}

	return httpClient
}

func getCertPool(skip bool, pemCerts []byte) (pool *x509.CertPool) {
	if !skip {
		pool = x509.NewCertPool()
		pool.AppendCertsFromPEM(pemCerts)
	}
	return
}

// RequestBuilder is a http request builder
type RequestBuilder interface {
	Build() (*http.Request, error)
	Method(method string) RequestBuilder
	Url(url string) RequestBuilder
	Body(body io.Reader) RequestBuilder
	AddParam(key, value string) RequestBuilder
	AddHeader(key, value string) RequestBuilder
	AddAuthorization(value string) RequestBuilder
}

type requestBuilder struct {
	method  string
	url     string
	body    io.Reader
	params  map[string]string
	headers map[string]string
}

// NewRequestBuilder returns a new request builder that adders the headers for application JSON and accepts */*
func NewRequestBuilder() RequestBuilder {
	return &requestBuilder{
		headers: map[string]string{
			"Accept":       "*/*",
			"Content-Type": "application/json",
		},
		params: map[string]string{},
	}
}

func (b *requestBuilder) Method(method string) RequestBuilder {
	b.method = method
	return b
}

func (b *requestBuilder) Url(url string) RequestBuilder {
	b.url = url
	return b
}

func (b *requestBuilder) Body(body io.Reader) RequestBuilder {
	b.body = body
	return b
}

func (b *requestBuilder) AddParam(key, value string) RequestBuilder {
	b.params[key] = value
	return b
}

func (b *requestBuilder) AddHeader(key, value string) RequestBuilder {
	b.headers[key] = value
	return b
}

func (b *requestBuilder) AddAuthorization(value string) RequestBuilder {
	b.headers["Authorization"] = value
	return b
}

func (b *requestBuilder) Build() (*http.Request, error) {
	request, err := http.NewRequest(b.method, b.url, b.body)
	if err == nil {
		addHeaders(request, b.headers)
		addParams(request, b.params)
	}

	return request, err
}

func addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func addParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
}
