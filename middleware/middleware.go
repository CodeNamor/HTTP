package middleware

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// AddContextValues is a common middleware which will populate the context with the fields from the context and log using said fields.
func AddContextValues(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
			r.Header.Set("X-Request-ID", reqID)
		}
		log.Debugf("incoming request %s %s %s", r.Method, r.RequestURI, reqID)
		handler.ServeHTTP(w, r)
	})
}
