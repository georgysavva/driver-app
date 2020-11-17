package httpmiddleware

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func NewLoggingMiddleware(next http.Handler, logger log.FieldLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{"method": r.Method, "path": r.RequestURI}).Info()
		next.ServeHTTP(w, r)
	})
}
