package gateway

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func logUnhandledError(logger log.FieldLogger, err error) {
	logger.WithError(err).Error("Unhandled error occurred")
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
