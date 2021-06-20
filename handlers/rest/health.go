package rest

import (
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
)

const healthTopic = "health"

type healthController struct {
	admin sarama.ClusterAdmin
	prod  sarama.SyncProducer
}

func newHealthController(saramaClient sarama.Client) (*healthController, error) {
	admin, err := sarama.NewClusterAdminFromClient(saramaClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init sarama admin")
	}

	prod, err := sarama.NewSyncProducerFromClient(saramaClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init sarama sync producer")
	}

	return &healthController{
		admin: admin,
		prod:  prod,
	}, nil
}

type status struct {
	Status string `json:"status"`
}

func fail(w http.ResponseWriter, r *http.Request, serviceName string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	render.JSON(w, r, status{Status: serviceName + ": " + err.Error()})
}

func ok(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, status{Status: "ok"})
}

func (hc *healthController) Readiness(w http.ResponseWriter, r *http.Request) {
	if err := hc.admin.CreateTopic(healthTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false); err != nil {
		fail(w, r, "kafka admin", err)
		return
	}

	if _, _, err := hc.prod.SendMessage(&sarama.ProducerMessage{Topic: healthTopic}); err != nil {
		fail(w, r, "kafka", err)
		return
	}

	if err := hc.admin.DeleteTopic(healthTopic); err != nil {
		fail(w, r, "kafka admin", err)
		return
	}

	ok(w, r)
}
