package zombiedriver

import (
	"context"
	"math"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
	"github.com/heetch/georgysavva-technical-test/zombie-driver/pkg/distance"
)

type Service interface {
	GetDriver(ctx context.Context, driverID string) (*Driver, error)
}

//go:generate mockery --name Service

type ZombiePredicate struct {
	DistanceThreshold int           `yaml:"distance_threshold"` // In meters.
	TimeInterval      time.Duration `yaml:"time_interval"`
}

type ServiceImpl struct {
	driverloc driverloc.GetterService
	logger    log.FieldLogger
	predicate *ZombiePredicate
}

func NewService(dl driverloc.GetterService, logger log.FieldLogger, predicate *ZombiePredicate,
) *ServiceImpl {
	return &ServiceImpl{
		driverloc: dl,
		logger:    logger,
		predicate: predicate,
	}
}

func (s *ServiceImpl) GetDriver(ctx context.Context, driverID string) (*Driver, error) {
	ctxLogger := s.logger.WithFields(log.Fields{
		"driver_id":     driverID,
		"time_interval": s.predicate.TimeInterval,
	})
	ctxLogger.Info("Request driver locations from the driver-location service")
	locations, err := s.driverloc.GetLocations(ctx, driverID, s.predicate.TimeInterval)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get driver locations from the driver-location service")
	}

	distanceDriven := calculateDistanceDriven(locations)
	isZombie := distanceDriven < s.predicate.DistanceThreshold
	ctxLogger.WithFields(log.Fields{
		"distance_driven":    distanceDriven,
		"distance_threshold": s.predicate.DistanceThreshold,
		"zombie":             isZombie,
	}).Info("Calculated distance driven for the driver")

	driver := &Driver{ID: driverID, IsZombie: isZombie}
	return driver, nil
}

func calculateDistanceDriven(locations []*driverloc.Location) int {
	if len(locations) <= 1 {
		return 0
	}
	var distanceDriven float64
	for i := 0; i < len(locations)-1; i++ {
		start, stop := locations[i], locations[i+1]
		locationsDistance := distance.Calculate(start.Latitude, start.Longitude, stop.Latitude, stop.Longitude)
		distanceDriven += locationsDistance
	}
	return int(math.Round(distanceDriven))
}
