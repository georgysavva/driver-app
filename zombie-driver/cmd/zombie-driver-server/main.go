package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/georgysavva/driver-app/driver-location/pkg/clients/driverlochttp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/georgysavva/driver-app/zombie-driver/pkg/config"
	"github.com/georgysavva/driver-app/zombie-driver/pkg/httpmiddleware"
	"github.com/georgysavva/driver-app/zombie-driver/pkg/zombiedriver"
)

// Improvement: allow to pass a custom config path.
const defaultConfigPath = "config.yaml"

func main() {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	conf, err := config.ParseConfig(defaultConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse config")
	}
	httpClient := http.DefaultClient

	driverLocationClient, err := driverlochttp.NewClient(httpClient, conf.DriverLocationService.BaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed initialize driver-location service http client")
	}

	service := zombiedriver.NewService(
		driverLocationClient, logger.WithField("component", "service"), conf.App.ZombiePredicate,
	)

	httpHandler := zombiedriver.MakeHTTPHandler(service, logger.WithField("component", "http-handler"))

	httpHandler = httpmiddleware.NewLoggingMiddleware(httpHandler, logger)
	httpServer := http.Server{Addr: fmt.Sprintf(":%d", conf.HTTPServer.Port), Handler: httpHandler}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("HTTP server unexpectedly stopped")
		}
	}()
	logger.WithField("http_port", conf.HTTPServer.Port).Info("HTTP server successfully started")

	terminationChan := make(chan os.Signal, 1)
	signal.Notify(terminationChan, syscall.SIGINT, syscall.SIGTERM)
	<-terminationChan

	logger.WithField("shutdown_timeout", conf.HTTPServer.ShutdownTimeout).Info("Stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), conf.HTTPServer.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Couldn't properly shutdown http server")
	}
	logger.Info("HTTP server was successfully shutdown")

	httpClient.CloseIdleConnections()
}

