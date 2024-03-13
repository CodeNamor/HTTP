package apiclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sethgrid/pester"
)

// RetryClient is an interface for http.RetryClient
type RetryClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewExtendedHTTPClient wraps an existing *http.RetryClient with retry logic
// The maximum number of attempts is configured in Config.MaxRetries.
// The retry logic uses an exponential backoff with jitter strategy.
// Retry attempts are logged to the warning level of the default logrus logger
func NewExtendedHTTPClient(maxRetries int, hc *http.Client) RetryClient {
	rc := pester.NewExtendedClient(hc)
	rc.MaxRetries = maxRetries
	rc.Backoff = pester.ExponentialJitterBackoff
	rc.KeepLog = false // must be false so LogHook can be used
	rc.LogHook = createLogHook(maxRetries)

	InstrumentedClient := &InstrumentedHttpClient{
		client: rc,
	}
	return InstrumentedClient
}

// InstrumentedHttpClient instruments the request, so we can determine the
// timings and whether a keep-alive client was used
type InstrumentedHttpClient struct {
	client RetryClient
}

func (ihc InstrumentedHttpClient) Do(req *http.Request) (*http.Response, error) {
	req, requestDone := InstrumentHTTPRequest(req)
	// we could call requestDone after downloading the content to get that timing
	// as well but that would require passing this down, so this is simpler
	defer requestDone()
	return ihc.client.Do(req)
}

// createLogHook creates a LogHook function which only logs the intermediate
// errors since the final one is returned and logged already
func createLogHook(maxRetries int) func(errEntry pester.ErrEntry) {
	return func(errEntry pester.ErrEntry) {
		// only log intermediate retries, last retry if error is already returned and logged
		if errEntry.Attempt < maxRetries {
			var origErr error
			if errEntry.Err != nil {
				origErr = errEntry.Err
			} else {
				origErr = fmt.Errorf("unknown error") // err was nil
			}
			_ = errors.Wrapf(origErr, "attempt:%d retrying", errEntry.Attempt)
		}
	}
}
