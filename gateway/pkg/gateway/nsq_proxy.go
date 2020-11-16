package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type NSQProducer interface {
	Publish(topic string, body []byte) error
}

//go:generate mockery --name NSQProducer

type NSQProxyFactory struct {
	producer NSQProducer
	logger   log.FieldLogger
}

func NewNSQProxyFactory(producer NSQProducer, logger log.FieldLogger) *NSQProxyFactory {
	return &NSQProxyFactory{
		producer: producer,
		logger:   logger,
	}
}

type NSQProxy struct {
	producer NSQProducer
	logger   log.FieldLogger
	conf     *NSQProxyConf
}

func (npf *NSQProxyFactory) NewProxy(conf *NSQProxyConf) (*NSQProxy, error) {
	if conf.Topic == "" {
		return nil, errors.New("NSQ proxy config has an empty topic")
	}
	return &NSQProxy{
		producer: npf.producer,
		logger:   npf.logger,
		conf:     conf,
	}, nil
}

type Message struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

func (np *NSQProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil && err != io.EOF {
		np.logger.WithError(err).Info("Can't decode request body into json")
		http.Error(w, errors.Wrap(err, "request body parsing failed").Error(), http.StatusBadRequest)
		return
	}
	msg := &Message{
		Command: np.conf.Message.Command,
		Data:    mergeRequestData(requestVars, requestData),
	}
	mesBody, err := json.Marshal(msg)
	if err != nil {
		logUnhandledError(np.logger, errors.Wrap(err, "can't encode nsq message body"))
		internalServerError(w)
		return
	}
	if err := np.producer.Publish(np.conf.Topic, mesBody); err != nil {
		logUnhandledError(np.logger, errors.Wrap(err, "failed to proxy message to nsq topic"))
		internalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "OK"); err != nil {
		logUnhandledError(np.logger, errors.Wrap(err, "failed to write 'OK' response to the client"))
		internalServerError(w)
		return
	}
}

func mergeRequestData(requestVars map[string]string, requestData map[string]interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{}, len(requestVars)+len(requestData))
	for k, v := range requestVars {
		resultMap[k] = v
	}
	for k, v := range requestData {
		resultMap[k] = v
	}
	return resultMap
}
