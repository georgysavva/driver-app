package driverloc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func MakeHTTPHandler(service GetterService, logger log.FieldLogger) http.Handler {
	router := mux.NewRouter()
	ha := &httpAPI{service: service, logger: logger}
	router.HandleFunc("/drivers/{id}/locations", ha.getLocations).Methods("GET")
	return router
}

type httpAPI struct {
	service GetterService
	logger  log.FieldLogger
}

func (ha *httpAPI) getLocations(w http.ResponseWriter, r *http.Request) {
	driverID := mux.Vars(r)["id"]
	ctxLogger := ha.logger.WithField("driver_id", driverID)
	timeIntervalMinutes, err := parseMinutesParam(r)
	if err != nil {
		ctxLogger.WithError(err).Info("Query params are invalid, return 400")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeInterval := time.Minute * time.Duration(timeIntervalMinutes)
	ctxLogger.WithField("time_interval", timeInterval).Info("Request driver locations from the service")
	locations, err := ha.service.GetLocations(r.Context(), driverID, timeInterval)
	if err != nil {
		ha.logUnhandledError(errors.Wrap(err, "failed to request locations from the service"))
		internalServerError(w)
		return
	}

	if err := returnJSONData(w, locations); err != nil {
		ha.logUnhandledError(err)
		internalServerError(w)
		return
	}
}

func parseMinutesParam(r *http.Request) (int, error) {
	minutesParam := r.URL.Query()["minutes"]
	if len(minutesParam) == 0 {
		return 0, errors.New("'minutes' query param is missing")
	}
	minutesValue, err := strconv.Atoi(minutesParam[0])
	if err != nil {
		return 0, errors.New("'minutes' query param must be a number")
	}
	return minutesValue, nil
}

func (ha *httpAPI) logUnhandledError(err error) {
	ha.logger.WithError(err).Error("Unhandled error occurred")
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func returnJSONData(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(obj)
	return errors.Wrap(err, "can not encode or write data into http response")
}
