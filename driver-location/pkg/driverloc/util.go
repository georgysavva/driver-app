package driverloc

import (
	log "github.com/sirupsen/logrus"
)

func logUnhandledError(logger log.FieldLogger, err error) {
	logger.WithError(err).Error("Unhandled error occurred")
}
