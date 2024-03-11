/*
Package server provides helpers for creating http server.

# Creating http.Server that has graceful shutdown enabled

Instead of using the net/http ListenAndServe just use this package's function and it will do the same thing but will also setup a graceful shutdown handler for Control-c and SIGTERM (which Kubernetes uses to shutdown pods).

	address := ":8000" // localhost:8000
	router := mux.NewRouter().StrictSlash(true)
	err := ListenAndServe(address, router)
	if err != nil {
	  logging.Error(errors.WithErrorAndCause(err, "HTTPServer exited with error"))
	}

# Creating Readiness and Liveness Probe Handlers

Using CreateAndHandleReadinessLiveness will create atomic handlers for readiness and liveness and will mount them on the router at the specified url paths. You can toggle the state of the two handlers by passing a boolean to returned readyFn and livenessFn to change the state. If state for a handler if false (default) then it will return 503 Service Unavailable. Once fn is called to update to true then handler will return 200 OK.

	router := mux.NewRouter().StrictSlash(true)
	readyFn, livenessFn := CreateAndHandleReadinessLiveness(router, "/ready", "/liveness")
	// if using config.New then all parsing and auth key checks are done
	// so can instantly make these both true
	// if other processing needed to be done then you can do that first
	// before calling readyFn
	// Handlers for readiness and liveness were created and mounted on
	// the router at the url paths provided
	readyFn(true)
	livenessFn(true)

# Creating Atomic Handlers

Normally you don't need to create atomic handlers yourself since they are created by CreateAndHandleReadinessLiveness but if you need to then
you can create as follows.

	initialState := false
	handlerFunc, UpdateFn := CreateAtomicHandler(initialState)
	// handler is ready to be mounted on router

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/mynewroute", handlerFunc)

	// before calling this handler would return 503 Service Unavailable
	UpdateFn(true) // now handler will return 200 OK
*/
package server
