package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/Shopify/sarama"
	votingpb "github.com/egsam98/voting/proto"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"google.golang.org/protobuf/proto"
)

var envs struct {
	Kafka struct {
		Addr  string `envconfig:"KAFKA_ADDR"`
		Topic string `envconfig:"KAFKA_TOPIC"`
	}
}

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := run(); err != nil {
		log.Fatal().Stack().Err(err).Msg("main: Fatal error")
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Msg("main: Read ENVs from .env file")
	}
	if err := envconfig.Process("", &envs); err != nil {
		return errors.Wrap(err, "failed to parse ENVs to struct")
	}

	log.Info().
		Interface("envs", envs).
		Msg("main: ENVs")

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true

	prod, err := sarama.NewAsyncProducer([]string{envs.Kafka.Addr}, cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to start producer on %q", envs.Kafka.Addr)
	}

	go func() {
		for err := range prod.Errors() {
			log.Error().Stack().Err(err).Msg("main: Producer error")
		}
	}()

	go func() {
		for msg := range prod.Successes() {
			be := msg.Value.(sarama.ByteEncoder)
			vote := &votingpb.Vote{}
			if err := proto.Unmarshal(be, vote); err != nil {
				log.Error().Stack().Err(err).Msgf("main: Failed to unmarshal protobuf to %T", vote)
				continue
			}

			log.Debug().
				Interface("msg", msg).
				Interface("vote", vote).
				Msg("main: Producer sent message")
		}
	}()

	var candidateID int64 = 0
	for {
		b, err := proto.Marshal(&votingpb.Vote{
			Voter:       &votingpb.Voter{ID: rand.Int63n(101)},
			CandidateId: candidateID,
		})
		if err != nil {
			return errors.Wrap(err, "faile to marshal vote")
		}

		prod.Input() <- &sarama.ProducerMessage{
			Topic: envs.Kafka.Topic,
			Value: sarama.ByteEncoder(b),
		}

		time.Sleep(2 * time.Second)
		candidateID++
	}
}
