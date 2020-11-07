package driverloc_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
)

const (
	defaultDriverID      = "foo"
	driverLocationsLimit = 3
)

var (
	baseTime = time.Date(2020, 11, 7, 00, 00, 00, 00, time.UTC)
	ctx      = context.Background()
)

func TestService_UpdateLocations(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(0 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.863921, Longitude: 2.349211},
			fakeCurrentTime: baseTime.Add(5 * time.Second),
		},
	})

	actual, err := fakeRedis.ZMembers(defaultDriverID)
	require.NoError(t, err)
	expected := []string{
		`{"latitude":48.864193,"longitude":2.350498,"updated_at":"2020-11-07T00:00:00Z"}`,
		`{"latitude":48.863921,"longitude":2.349211,"updated_at":"2020-11-07T00:00:05Z"}`,
	}
	assert.Equal(t, expected, actual)
}

func TestService_UpdateLocations_OlderLocationsAreCleaned(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(0 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.863921, Longitude: 2.349211},
			fakeCurrentTime: baseTime.Add(5 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.862921, Longitude: 2.348211},
			fakeCurrentTime: baseTime.Add(10 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.861921, Longitude: 2.347211},
			fakeCurrentTime: baseTime.Add(15 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.860921, Longitude: 2.346211},
			fakeCurrentTime: baseTime.Add(20 * time.Second),
		},
	})

	actual, err := fakeRedis.ZMembers(defaultDriverID)
	require.NoError(t, err)
	expected := []string{
		`{"latitude":48.862921,"longitude":2.348211,"updated_at":"2020-11-07T00:00:10Z"}`,
		`{"latitude":48.861921,"longitude":2.347211,"updated_at":"2020-11-07T00:00:15Z"}`,
		`{"latitude":48.860921,"longitude":2.346211,"updated_at":"2020-11-07T00:00:20Z"}`,
	}
	assert.Equal(t, expected, actual)
}

func TestService_UpdateLocations_DuplicateCoordinatesAtDifferentTimes(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(0 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(5 * time.Second),
		},
	})

	actual, err := fakeRedis.ZMembers(defaultDriverID)
	require.NoError(t, err)
	expected := []string{
		`{"latitude":48.864193,"longitude":2.350498,"updated_at":"2020-11-07T00:00:00Z"}`,
		`{"latitude":48.864193,"longitude":2.350498,"updated_at":"2020-11-07T00:00:05Z"}`,
	}
	assert.Equal(t, expected, actual)
}

func TestService_UpdateLocations_DuplicateCoordinatesAtTheSameTime(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime,
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime,
		},
	})

	actual, err := fakeRedis.ZMembers(defaultDriverID)
	require.NoError(t, err)
	expected := []string{
		`{"latitude":48.864193,"longitude":2.350498,"updated_at":"2020-11-07T00:00:00Z"}`,
	}
	assert.Equal(t, expected, actual)
}

func TestService_UpdateLocations_DifferentCoordinatesAtTheSameTime(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime,
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.863921, Longitude: 2.349211},
			fakeCurrentTime: baseTime,
		},
	})

	actual, err := fakeRedis.ZMembers(defaultDriverID)
	require.NoError(t, err)
	expected := []string{
		`{"latitude":48.863921,"longitude":2.349211,"updated_at":"2020-11-07T00:00:00Z"}`,
		`{"latitude":48.864193,"longitude":2.350498,"updated_at":"2020-11-07T00:00:00Z"}`,
	}
	assert.Equal(t, expected, actual)
}

func TestService_GetLocations(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(0 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.863921, Longitude: 2.349211},
			fakeCurrentTime: baseTime.Add(5 * time.Second),
		},
		{
			coords:          &driverloc.Coordinates{Latitude: 48.862921, Longitude: 2.348211},
			fakeCurrentTime: baseTime.Add(10 * time.Second),
		},
	})

	service.SetTimeNowFn(func() time.Time {
		return baseTime.Add(15 * time.Second)
	})
	timeInterval := 10 * time.Second
	actual, err := service.GetLocations(ctx, defaultDriverID, timeInterval)
	require.NoError(t, err)

	expected := []*driverloc.Location{
		{
			Coordinates: &driverloc.Coordinates{Latitude: 48.863921, Longitude: 2.349211},
			Time:        baseTime.Add(5 * time.Second),
		},
		{
			Coordinates: &driverloc.Coordinates{Latitude: 48.862921, Longitude: 2.348211},
			Time:        baseTime.Add(10 * time.Second),
		},
	}
	assert.Equal(t, expected, actual)
}

func TestService_GetLocations_NoLocationsInTheTimeInterval(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	insertLocations(t, service, defaultDriverID, []*toInsert{
		{
			coords:          &driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
			fakeCurrentTime: baseTime.Add(0 * time.Second),
		},
	})

	service.SetTimeNowFn(func() time.Time {
		return baseTime.Add(15 * time.Second)
	})
	timeInterval := 10 * time.Second
	actual, err := service.GetLocations(ctx, defaultDriverID, timeInterval)
	require.NoError(t, err)

	assert.Empty(t, actual)
}

func TestService_GetLocations_NoLocationsAtAll(t *testing.T) {
	t.Parallel()
	service, fakeRedis := setup(t)
	defer fakeRedis.Close()

	service.SetTimeNowFn(func() time.Time {
		return baseTime.Add(15 * time.Second)
	})
	timeInterval := 10 * time.Second
	actual, err := service.GetLocations(ctx, defaultDriverID, timeInterval)
	require.NoError(t, err)

	assert.Empty(t, actual)
}

func setup(t *testing.T) (*driverloc.ServiceImpl, *miniredis.Miniredis) {
	t.Helper()
	fakeRedis, err := miniredis.Run()
	require.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: fakeRedis.Addr()})
	logger := log.New()
	logger.SetLevel(log.ErrorLevel)
	s := driverloc.NewService(redisClient, logger, driverLocationsLimit)
	return s, fakeRedis
}

type toInsert struct {
	coords          *driverloc.Coordinates
	fakeCurrentTime time.Time
}

func insertLocations(t *testing.T, s *driverloc.ServiceImpl, driverID string, inserts []*toInsert) {
	t.Helper()
	for _, insert := range inserts {
		s.SetTimeNowFn(func() time.Time {
			return insert.fakeCurrentTime
		})
		err := s.UpdateLocations(ctx, driverID, insert.coords)
		require.NoError(t, err)
	}
}
