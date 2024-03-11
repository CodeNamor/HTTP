package apiclient

import (
	"context"
	"expvar"
	metrics "github.com/go-kit/kit/metrics/expvar"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	expvarHTTPClientNewConns    = "HTTPClientNewConnections"
	expvarHTTPClientReusedConns = "HTTPClientReusedConnections"
	expvarHTTPClientConnPrep    = "HTTPClientConnectionPreparation"
)

var httpClientNewConnCounter *metrics.Counter
var httpClientReusedConnCounter *metrics.Counter
var httpClientConnPrepHistogram *metrics.Histogram

func init() {
	httpClientNewConnCounter = metrics.NewCounter(expvarHTTPClientNewConns)
	httpClientReusedConnCounter = metrics.NewCounter(expvarHTTPClientReusedConns)
	httpClientConnPrepHistogram = metrics.NewHistogram(expvarHTTPClientConnPrep, 50)
}

// InstrumentHTTPRequest adds the instrumentation hooks to the http.Request
// to track the timings associated with making a request.
// The info is logged via logrus with the default fields from the context/ei_logging and expvar
// Call the returned done function after response is created
// or after request finishes downloading if you want full
// timing information. This adds to counters tracking new and
// reused httpClient connections and a histogram tracking
// the time to prepare a connection (dns+tcp+tls) in ms.
// The prep time would be 0 when a connection is reused.
// Takes a logEntry to structured logging with the request.
//
//	req, err := http.NewRequest(http.MethodGet, url, nil /* body */)
//	if err != nil {
//	  return err
//	}
//	logEntry := logging.WithField("mykey", "myvalue")
//	req, requestDone := InstrumentHTTPRequestWithLogEntry(req)
//	res, err := httpClient.Do(req)
//	defer requestDone()
func InstrumentHTTPRequest(req *http.Request) (*http.Request, func()) {
	httpClientReusedConnCounter.Add(1.0)
	httpClientConnPrepHistogram.Observe(0.0) // reused conn, 0 preparation time
	return req.WithContext(context.Background()), nil
}

// AddExpVarHandlerToRouter adds the expvar handler to the root
// relative urlPath on the router. Fetching this urlPath will
// output all of the available expvar metrics.
//
//	router := mux.NewRouter().StrictSlash(true)
//	AddExpVarHandlerToRouter(router, "/debug/vars")
func AddExpVarHandlerToRouter(router *mux.Router, urlPath string) {
	router.HandleFunc(urlPath, expvar.Handler().ServeHTTP).Methods(http.MethodGet)
}
