package driverloc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type UpdaterService interface {
	UpdateLocations(ctx context.Context, driverID string, coordinates *Coordinates) error
}

//go:generate mockery --name UpdaterService

type GetterService interface {
	GetLocations(ctx context.Context, driverID string, timeInterval time.Duration) ([]*Location, error)
}

//go:generate mockery --name GetterService

type Service interface {
	UpdaterService
	GetterService
}

// Improvement: move redis interaction to a separate storage layer and use an abstraction in service.
type ServiceImpl struct {
	redis                *redis.Client
	logger               log.FieldLogger
	driverLocationsLimit int
	timeNowFn            func() time.Time
}

func NewService(r *redis.Client, logger log.FieldLogger, driverLocationsLimit int) *ServiceImpl {
	return &ServiceImpl{
		redis:                r,
		logger:               logger,
		driverLocationsLimit: driverLocationsLimit,
		timeNowFn:            time.Now,
	}
}

func (s *ServiceImpl) UpdateLocations(ctx context.Context, driverID string, coordinates *Coordinates) error {
	now := s.timeNowFn().UTC()
	loc := &Location{Coordinates: coordinates, Time: now}
	locationData, err := json.Marshal(loc)
	if err != nil {
		return errors.Wrap(err, "failed to encode location data into json")
	}

	redisSetMember := &redis.Z{
		Score:  float64(now.Unix()),
		Member: string(locationData),
	}
	ctxLogger := s.logger.WithField("driver_id", driverID)
	ctxLogger.WithField("location", loc).Info("Save new driver location into Redis")
	if err := s.redis.ZAdd(ctx, driverID, redisSetMember).Err(); err != nil {
		return errors.Wrap(err, "failed to save new driver location into Redis")
	}

	ctxLogger.Info("Clean old driver locations in Redis")
	cleanedNum, err := s.redis.ZRemRangeByRank(ctx, driverID, 0, -1-int64(s.driverLocationsLimit)).Result()
	if err != nil {
		return errors.Wrap(err, "failed to clean old driver locations in Redis")
	}
	ctxLogger.WithField("cleaned_num", cleanedNum).Info("Cleaned old driver locations in Redis")

	return nil
}

func (s *ServiceImpl) GetLocations(ctx context.Context, driverID string, timeInterval time.Duration) (
	[]*Location, error) {
	now := s.timeNowFn().UTC()
	minScore := timeToRedisScore(now.Add(-timeInterval))
	redisRange := &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: "+inf",
	}
	ctxLogger := s.logger.WithField("driver_id", driverID)
	ctxLogger.WithFields(log.Fields{
		"min_time": redisRange.Min,
		"max_time": redisRange.Max,
	}).Info("Get driver locations from Redis by time range")
	locationsData, err := s.redis.ZRangeByScore(ctx, driverID, redisRange).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get driver locations from Redis")
	}
	ctxLogger.WithField("locations_num", len(locationsData)).Info("Retrieved driver locations from Redis")

	locations, err := decodeLocations(locationsData)
	return locations, errors.WithStack(err)
}

func decodeLocations(locationsData []string) ([]*Location, error) {
	locations := make([]*Location, len(locationsData))
	for i, data := range locationsData {
		loc := &Location{}
		if err := json.Unmarshal([]byte(data), loc); err != nil {
			return nil, errors.Wrapf(err, "can't decode location data %s", data)
		}
		locations[i] = loc
	}
	return locations, nil
}

func timeToRedisScore(t time.Time) float64 {
	return float64(t.Unix())
}
