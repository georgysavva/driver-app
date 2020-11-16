package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/config"
	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
)

// Improvement: Allow to pass a custom config path.
const defaultConfigPath = "config.yaml"

func main() {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	conf, err := config.ParseConfig(defaultConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse config")
	}
	redisClient := redis.NewClient(&redis.Options{Addr: conf.Redis.Address})
	defer redisClient.Close()
	service := driverloc.NewService(redisClient, logger.WithField("component", "service"), conf.App.DriverLocationsLimit)

	// HTTP server
	httpHandler := driverloc.MakeHTTPHandler(service, logger.WithField("component", "http-handler"))
	httpServer := http.Server{Addr: fmt.Sprintf(":%d", conf.HTTPServer.Port), Handler: httpHandler}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("HTTP server unexpectedly stopped")
		}
	}()
	logger.WithField("http_port", conf.HTTPServer.Port).Info("HTTP server successfully started")

	// NSQ consumer
	nsqHandler := driverloc.NewNSQHandler(service, logger.WithField("component", "nsq-handler"))
	nsqConsumer, err := nsq.NewConsumer(conf.NSQ.Topic, conf.NSQ.Channel, nsq.NewConfig())
	if err != nil {
		logger.WithError(err).Fatal("Couldn't initialize nsq consumer")
	}
	nsqConsumer.AddConcurrentHandlers(nsqHandler, conf.NSQ.WorkersNum)
	if err := nsqConsumer.ConnectToNSQLookupds(conf.NSQ.LookupdAddresses); err != nil {
		logger.WithError(err).Fatal("Couldn't connect nsq consumer to nsqlookupds")
	}
	logger.Info("NSQ consumer successfully started")

	terminationChan := make(chan os.Signal, 1)
	signal.Notify(terminationChan, syscall.SIGINT, syscall.SIGTERM)
	<-terminationChan

	logger.Info("Stopping NSQ consumer")
	nsqConsumer.Stop()
	logger.Info("NSQ consumer stopped")

	logger.WithField("shutdown_timeout", conf.HTTPServer.ShutdownTimeout).Info("Stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), conf.HTTPServer.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Couldn't properly shutdown http server")
	}
	logger.Info("HTTP server was successfully shutdown")

	if err := redisClient.Close(); err != nil {
		logger.WithError(err).Fatal("Couldn't close redis client")
	}
}
