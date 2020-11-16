package zombiedriver_test

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc/mocks"
	"github.com/heetch/georgysavva-technical-test/zombie-driver/pkg/zombiedriver"
)

const defaultDriverID = "foo"

func TestService_GetDriver(t *testing.T) {
	t.Parallel()
	driverlocMock := &mocks.GetterService{}
	logger := log.New()
	logger.SetLevel(log.ErrorLevel)
	timeInterval := 5 * time.Minute
	baseTime := time.Now()
	driverlocMock.On("GetLocations",
		mock.MatchedBy(func(_ context.Context) bool { return true }), // anything of type context.Context
		defaultDriverID, timeInterval,
	).Return([]*driverloc.Location{
		{
			Coordinates: &driverloc.Coordinates{
				Latitude:  48.864193,
				Longitude: 2.350498,
			},
			Time: baseTime.Add(-15 * time.Second),
		},
		{
			Coordinates: &driverloc.Coordinates{
				Latitude:  48.863193,
				Longitude: 2.351498,
			},
			Time: baseTime.Add(-10 * time.Second),
		},
		{
			Coordinates: &driverloc.Coordinates{
				Latitude:  48.862193,
				Longitude: 2.352498,
			},
			Time: baseTime.Add(-5 * time.Second),
		},
	}, nil)

	cases := []struct {
		name              string
		distanceThreshold int
		isZombie          bool
	}{
		{
			name:              "zombie",
			distanceThreshold: 500,
			isZombie:          true,
		},
		{
			name:              "not zombie",
			distanceThreshold: 200,
			isZombie:          false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := zombiedriver.NewService(driverlocMock, logger, &zombiedriver.ZombiePredicate{
				DistanceThreshold: tc.distanceThreshold,
				TimeInterval:      timeInterval,
			})

			actual, err := service.GetDriver(context.Background(), defaultDriverID)
			require.NoError(t, err)

			assert.Equal(t, &zombiedriver.Driver{ID: defaultDriverID, IsZombie: tc.isZombie}, actual)
		})
	}
}
