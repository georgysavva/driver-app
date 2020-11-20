package driverloc_test

import (
	"context"
	"testing"

	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/georgysavva/driver-app/driver-location/pkg/driverloc"
	"github.com/georgysavva/driver-app/driver-location/pkg/driverloc/mocks"
)

func TestNSQHandler_HandleMessage(t *testing.T) {
	t.Parallel()
	serviceMock := &mocks.UpdaterService{}
	serviceMock.On(
		"UpdateLocations",
		mock.MatchedBy(func(_ context.Context) bool { return true }), // match anything of type context.Context
		defaultDriverID,
		&driverloc.Coordinates{Latitude: 48.864193, Longitude: 2.350498},
	).Return(nil)
	nsqHandler := newNSQHandler(serviceMock)

	body := `
	{
		"command": "update-driver-locations",
		"data": {
			"id": "foo",
			"latitude": 48.864193,
			"longitude": 2.350498
		}
	}`
	msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4}, []byte(body))
	err := nsqHandler.HandleMessage(msg)
	require.NoError(t, err)

	serviceMock.AssertExpectations(t)
}

func TestNSQHandler_HandleMessage_RequestError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
	}{
		{
			name: "empty body",
			body: ``,
		},
		{
			name: "body is not a json",
			body: `foo`,
		},
		{
			name: "body json structure is invalid",
			body: `{"command": 5}`,
		},
		{
			name: "unsupported command",
			body: `{"command": "foo"}`,
		},
		{
			name: "update driver locations driver_id is missing",
			body: `
			{
				"command": "update-driver-locations",
				"data": {
					"latitude": 48.864193,
					"longitude": 2.350498
				}
			}`,
		},
		{
			name: "update driver locations latitude is missing",
			body: `
			{
				"command": "update-driver-locations",
				"data": {
					"id": "foo",
					"longitude": 2.350498
				}
			}`,
		},
		{
			name: "update driver location longitude is missing",
			body: `
			{
				"command": "update-driver-locations",
				"data": {
					"id": "foo",
					"latitude": 48.864193,
				}
			}`,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			serviceMock := &mocks.UpdaterService{}
			serviceMock.On(
				"UpdateLocations",
				mock.Anything /* ctx */, mock.Anything /* driverID */, mock.Anything, /* coordinates */
			).Return(nil)

			nsqHandler := newNSQHandler(serviceMock)
			msg := newNSQMessage(tc.body)
			err := nsqHandler.HandleMessage(msg)
			require.NoError(t, err)

			serviceMock.AssertNumberOfCalls(t, "UpdateLocations", 0)
		})
	}
}

func newNSQMessage(body string) *nsq.Message {
	return nsq.NewMessage(nsq.MessageID{1, 2, 3, 4}, []byte(body))
}

func newNSQHandler(serviceMock *mocks.UpdaterService) *driverloc.NSQHandler {
	logger := log.New()
	logger.SetLevel(log.ErrorLevel)
	return driverloc.NewNSQHandler(serviceMock, logger)
}
