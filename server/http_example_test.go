package server

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func ExampleListenAndServe() {
	startServerFn := func() {
		address := ":8000" // localhost:8000
		router := mux.NewRouter().StrictSlash(true)
		err := ListenAndServe(address, router)
		if err != nil {
			log.Errorf("HTTPServer exited with error: %v", err.Error())
		}
	}
	// if running for real, call the fn as below
	// startServerFn()

	// this line is just to make compiler happy since not calling fn
	fmt.Println(reflect.TypeOf(startServerFn).Kind().String())
}

func ExampleCreateAndHandleReadinessLiveness() {
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
}

func ExampleCreateAtomicHandler() {
	// Normally you don't need to call this if just using
	// CreateAndHandleReadinessLiveness but if you want to do this manually
	handlerFunc, updateFn := CreateAtomicHandler(false)
	// handler is ready to be mounted on router

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/mynewroute", handlerFunc)

	// before calling this handler would return 503 Service Unavailable
	updateFn(true) // now handler will return 200 OK
}
