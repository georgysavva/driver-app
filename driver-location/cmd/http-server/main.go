package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/config"
	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
)

func main() {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	conf, err := config.ParseConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse config")
	}
	redisClient := redis.NewClient(&redis.Options{Addr: conf.Redis.Address})
	defer redisClient.Close()
	service := driverloc.NewService(redisClient, logger.WithField("component", "service"), conf.App.DriverLocationsLimit)
	handler := driverloc.MakeHTTPHandler(service, logger.WithField("component", "http-handler"))
	srv := http.Server{Addr: fmt.Sprintf(":%d", conf.HTTPServer.Port), Handler: handler}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.WithError(err).Fatal("HTTP server unexpectedly stopped")
		}
	}()
	logger.WithField("port", conf.HTTPServer.Port).Info("Starting http server")

	<-done
	logger.WithField("shutdown_timeout", conf.HTTPServer.ShutdownTimeout).Info("Stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), conf.HTTPServer.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Couldn't properly shutdown http server")
	}
	logger.Info("HTTP server was successfully shutdown")

	if err := redisClient.Close(); err != nil {
		logger.WithError(err).Fatal("Couldn't close redis client")
	}
}
