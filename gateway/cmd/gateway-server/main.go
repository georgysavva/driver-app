package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"

	"github.com/heetch/georgysavva-technical-test/gateway/pkg/config"
	"github.com/heetch/georgysavva-technical-test/gateway/pkg/gateway"
	"github.com/heetch/georgysavva-technical-test/gateway/pkg/httpmiddleware"
)

// Improvement: allow to pass a custom config path.
const defaultConfigPath = "config.yaml"

func main() {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	conf, err := config.ParseConfig(defaultConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse config")
	}
	nsqProducer, err := nsq.NewProducer(conf.NSQ.DaemonAddress, nsq.NewConfig())
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect nsq producer to daemon")
	}
	nsqProxyFactory := gateway.NewNSQProxyFactory(nsqProducer, logger.WithField("component", "nsq-proxy"))

	gatewayHandler, err := gateway.NewGateway(nsqProxyFactory, conf.URLs)
	if err != nil {
		logger.WithError(err).Fatal("Couldn't setup gateway handler")
	}

	gatewayHandler = httpmiddleware.NewLoggingMiddleware(gatewayHandler, logger)
	httpServer := http.Server{Addr: fmt.Sprintf(":%d", conf.HTTPServer.Port), Handler: gatewayHandler}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	logger.Info("Stopping NSQ producer")
	nsqProducer.Stop()
	logger.Info("NSQ producer stopped")
}
