package driverloc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const commandUpdateDriverLocations = "update-driver-locations"

type NSQHandler struct {
	service UpdaterService
	logger  log.FieldLogger
}

func NewNSQHandler(service UpdaterService, logger log.FieldLogger) *NSQHandler {
	return &NSQHandler{service: service, logger: logger}
}

type nsqRequest struct {
	Command string `json:"command"`
	Data    struct {
		DriverID  *string  `json:"id"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	} `json:"data"`
}

func (nh *NSQHandler) HandleMessage(m *nsq.Message) error {
	ctx := context.Background()
	ctxLogger := nh.logger.WithField("message_id", fmt.Sprintf("%s", m.ID))
	ctxLogger.WithField("message_body", string(m.Body)).Info("Received a new message")

	req, err := parseNSQRequest(m)
	if err != nil {
		ctxLogger.WithError(err).Info("Couldn't parse nsq message, finish processing")
		return nil
	}

	ctxLogger = ctxLogger.WithField("command", req.Command)
	if req.Command != commandUpdateDriverLocations {
		ctxLogger.Info("NSQ request contains unsupported command")
		return nil
	}

	ctxLogger.Info("Handle nsq request")
	if err := nh.updateLocations(ctx, req); err != nil {
		logUnhandledError(ctxLogger, err)
		return err
	}

	return nil
}

func (nh *NSQHandler) updateLocations(ctx context.Context, req *nsqRequest) error {
	data := req.Data
	if data.DriverID == nil || data.Latitude == nil || data.Longitude == nil {
		nh.logger.Info("NSQ request data is incomplete: " +
			"'driver_id', 'latitude', 'longitude' fields must be set, finish_processing")
		return nil
	}
	coordinates := &Coordinates{
		Latitude:  *data.Latitude,
		Longitude: *data.Longitude,
	}
	nh.logger.WithField("driver_id", data.DriverID).Info("Call service to update driver locations")
	err := nh.service.UpdateLocations(ctx, *data.DriverID, coordinates)
	return errors.Wrap(err, "failed to call service to update driver locations")
}

func parseNSQRequest(m *nsq.Message) (*nsqRequest, error) {
	if len(m.Body) == 0 {
		return nil, errors.New("message has an empty body")
	}
	req := &nsqRequest{}
	if err := json.Unmarshal(m.Body, req); err != nil {
		return nil, errors.Wrap(err, "cannot decode message body into as request struct")
	}
	return req, nil
}
