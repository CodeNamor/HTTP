package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListenAndServe(t *testing.T) {
	// only testable on non-windows

	address := ":8000"
	testcases := []struct {
		name             string
		signal           os.Signal
		expectedErrorStr string
	}{
		{
			name:             "SIGINT Ctrl+C",
			signal:           os.Interrupt,
			expectedErrorStr: "",
		},
		{
			name:             "SIGTERM Kubernetes shutdown",
			signal:           syscall.SIGTERM,
			expectedErrorStr: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pid := os.Getpid()
			process, err := os.FindProcess(pid)
			require.NoError(t, err)
			router := mux.NewRouter().StrictSlash(true)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				// immediately after listening signal this asyncrhonously
				go func() {
					time.Sleep(3 * time.Second) // sleep to allow server to start
					err := process.Signal(tc.signal)
					require.NoError(t, err)
				}()
				err = ListenAndServe(address, router)
			}()
			wg.Wait()
			if tc.expectedErrorStr == "" {
				require.NoError(t, err)
			} else { // expecting error
				require.Error(t, err)
				require.Contains(t, err, tc.expectedErrorStr)
			}
		})
	}
}

func Test_CreateAndHandleReadinessLiveness(t *testing.T) {
	readyURL := "/ready"
	liveURL := "/live"
	testcases := []struct {
		name                 string
		readyInitial         bool
		liveInitial          bool
		readyNext            bool
		liveNext             bool
		expectedInitialReady string
		expectedInitialLive  string
		expectedNextReady    string
		expectedNextLive     string
	}{
		{
			name:                 "0 - initially all false, stays false",
			readyInitial:         false,
			liveInitial:          false,
			readyNext:            false,
			liveNext:             false,
			expectedInitialReady: "Service Unavailable",
			expectedInitialLive:  "Service Unavailable",
			expectedNextReady:    "Service Unavailable",
			expectedNextLive:     "Service Unavailable",
		},
		{
			name:                 "1 - initially all false, liveness becomes true",
			readyInitial:         false,
			liveInitial:          false,
			readyNext:            false,
			liveNext:             true,
			expectedInitialReady: "Service Unavailable",
			expectedInitialLive:  "Service Unavailable",
			expectedNextReady:    "Service Unavailable",
			expectedNextLive:     "OK",
		},
		{
			name:                 "2 - initially all false, ready becomes true",
			readyInitial:         false,
			liveInitial:          false,
			readyNext:            true,
			liveNext:             false,
			expectedInitialReady: "Service Unavailable",
			expectedInitialLive:  "Service Unavailable",
			expectedNextReady:    "OK",
			expectedNextLive:     "Service Unavailable",
		},
		{
			name:                 "3 - initially all false, both become true",
			readyInitial:         false,
			liveInitial:          false,
			readyNext:            true,
			liveNext:             true,
			expectedInitialReady: "Service Unavailable",
			expectedInitialLive:  "Service Unavailable",
			expectedNextReady:    "OK",
			expectedNextLive:     "OK",
		},
		{
			name:                 "4 - initially all true, both become false",
			readyInitial:         true,
			liveInitial:          true,
			readyNext:            false,
			liveNext:             false,
			expectedInitialReady: "OK",
			expectedInitialLive:  "OK",
			expectedNextReady:    "Service Unavailable",
			expectedNextLive:     "Service Unavailable",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			router := mux.NewRouter().StrictSlash(true)
			readyFn, liveFn := CreateAndHandleReadinessLiveness(router, readyURL, liveURL)
			readyFn(tc.readyInitial)
			liveFn(tc.liveInitial)

			// ready initial
			req, _ := http.NewRequest("GET", readyURL, nil)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Contains(t, res.Body.String(), tc.expectedInitialReady)

			// live initial
			req, _ = http.NewRequest("GET", liveURL, nil)
			res = httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Contains(t, res.Body.String(), tc.expectedInitialLive)

			readyFn(tc.readyNext)
			liveFn(tc.liveNext)

			// ready next
			req, _ = http.NewRequest("GET", readyURL, nil)
			res = httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Contains(t, res.Body.String(), tc.expectedNextReady)

			// live next
			req, _ = http.NewRequest("GET", liveURL, nil)
			res = httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Contains(t, res.Body.String(), tc.expectedNextLive)
		})
	}
}

func Test_isNormalShutdown(t *testing.T) {
	testcases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil err",
			err:      nil,
			expected: false,
		},
		{
			name:     "http.ErrServerClosed",
			err:      http.ErrServerClosed,
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("my error"),
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, isNormalShutdown(tc.err))
		})
	}
}

func Test_DefaultHandler(t *testing.T) {

	expected := ""

	fn := func(rw http.ResponseWriter, req *http.Request) {
		contentTypeHeader := rw.Header().Get("Content-Type")
		assert.Equal(t, expected, contentTypeHeader)
	}

	responseRecorder := httptest.NewRecorder()
	mockRequest := httptest.NewRequest("GET", "https://centene.com", nil)

	fn(responseRecorder, mockRequest)

	expected = "application/json"

	handlerFn := DefaultHandler(fn)

	handlerFn(responseRecorder, mockRequest)
}
