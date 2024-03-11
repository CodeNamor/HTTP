package server

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ListenAndServe creates a http.Server, sets up signal handler
// to listen for SIGINT and SIGTERM to perform graceful shutdown,
// and launches the server. If the server returns anything other
// than a normal close, then it is returned, otherwise returns nil.ListenAndServe
// It logs at the Info level.
func ListenAndServe(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	signals := make(chan os.Signal, 1)                    // channel to listen to signals
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM) // forward ctrl-c + SIGTERM to signals chan

	wgServer := sync.WaitGroup{}
	wgServer.Add(2) // wait for signal and listen routines

	var serverError error

	go func() { // routine to wait for any shutdown signal
		defer wgServer.Done()
		killSignal := <-signals // wait for a signal
		switch killSignal {
		case os.Interrupt:
		case syscall.SIGTERM:
			log.Info("SIGTERM received (Kubernetes shutdown?)")
		case nil: // if we sent a nil signal in, then exit now
			return // exit now, probably due to server error
		}

		log.Info("graceful shutdown initiated...")
		err := srv.Shutdown(context.Background())
		if err != nil {
			log.Errorf("graceful shutdown errored: %v", err)
		} else {
			log.Info("graceful shutdown complete")
		}
	}()

	go func() { // routine to begin listening and serving
		defer wgServer.Done()
		serverError = srv.ListenAndServe()
		if !isNormalShutdown(serverError) {
			signals <- nil // we need to exit now, tell other routine
		}
	}()

	// wait for routines to finish
	wgServer.Wait()
	if isNormalShutdown(serverError) {
		return nil
	}

	return serverError
}

// CreateAndHandleReadinessLiveness creates atomic handlers for
// readiness and liveness, it registers them at the readyURLPath
// and liveURLPath, and it returns a readyFn and livenessFn updater
// functions which can be used to set the readiness and liveness
// status at any time. The atomicHandlers will return 200 OK or
// 503 Service Unavailable based on the state of the atomicHandler
// from the last update using the UpdateFn (readyFn, livenessFn),
// the default state is false for both.
func CreateAndHandleReadinessLiveness(router *mux.Router, readyURLPath string, liveURLPath string) (UpdateFn, UpdateFn) {
	//set to false initially.
	readyHandler, readyFn := CreateAtomicHandler(false)
	router.HandleFunc(readyURLPath, readyHandler)

	//set to false initially.
	livenessHandler, livenessFn := CreateAtomicHandler(false)
	router.HandleFunc(liveURLPath, livenessHandler)

	return readyFn, livenessFn // return updater fns
}

// UpdateFn is function variable. used as a return value of CreateAndHandleReadinessLiveness
type UpdateFn func(bool)

// CreateAtomicHandler creates an atomic value handler which
// returns 200 OK or 503 Service Unavailable based on the state
// of the atomic value which defaults to initialValue.
// It returns the handler and an update func which can be used
// to atomically update the state.
func CreateAtomicHandler(initialValue bool) (http.HandlerFunc, UpdateFn) {
	atomicValue := &atomic.Value{}
	atomicValue.Store(initialValue)

	handler := func(w http.ResponseWriter, _ *http.Request) {
		if !atomicValue.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("OK"))
	}

	updateFn := func(value bool) {
		atomicValue.Store(value)
	}

	return handler, updateFn
}

func isNormalShutdown(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}

// DefaultHandler sets the Content-Type header to "application/json"
func DefaultHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		fn(rw, req)
	}
}
