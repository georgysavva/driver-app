package zombiedriver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func MakeHTTPHandler(service Service, logger log.FieldLogger) http.Handler {
	router := mux.NewRouter()
	ha := &httpAPI{service: service, logger: logger}
	router.HandleFunc("/drivers/{id}", ha.getDriver).Methods("GET")
	return router
}

type httpAPI struct {
	service Service
	logger  log.FieldLogger
}

func (ha *httpAPI) getDriver(w http.ResponseWriter, r *http.Request) {
	driverID := mux.Vars(r)["id"]
	ctxLogger := ha.logger.WithField("driver_id", driverID)

	ctxLogger.Info("Request driver from the service")
	driver, err := ha.service.GetDriver(r.Context(), driverID)
	if err != nil {
		logUnhandledError(ctxLogger, errors.Wrap(err, "failed to request driver from the service"))
		internalServerError(w)
		return
	}

	if err := returnJSONData(w, driver); err != nil {
		logUnhandledError(ctxLogger, err)
		internalServerError(w)
		return
	}
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func returnJSONData(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(obj)
	return errors.Wrap(err, "can not encode or write data into http response")
}

func logUnhandledError(logger log.FieldLogger, err error) {
	logger.WithError(err).Error("Unhandled error occurred")
}
