package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/egsam98/voting/parser/handlers/rest"
	"github.com/egsam98/voting/parser/services/votes"
)

var envs struct {
	Web struct {
		Addr            string        `envconfig:"WED_ADDR" default:"localhost:3000"`
		ShutdownTimeout time.Duration `envconfig:"WEB_SHUTDOWN_TIMEOUT" default:"5s"`
	}
	Kafka struct {
		Addr           string `envconfig:"KAFKA_ADDR"`
		ValidatorTopic string `envconfig:"KAFKA_VALIDATOR_TOPIC"`
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

	prod, err := sarama.NewSyncProducer([]string{envs.Kafka.Addr}, cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to start producer on %q", envs.Kafka.Addr)
	}

	defer func() {
		if err := prod.Close(); err != nil {
			log.Error().Stack().Err(err).Msg("main: Failed to close sarama producer")
		}
	}()

	srv := http.Server{
		Addr:    envs.Web.Addr,
		Handler: rest.API(votes.NewService(prod, envs.Kafka.ValidatorTopic)),
	}

	apiErr := make(chan error)
	go func() {
		log.Info().Msgf("main: Parser REST service is listening on %q", envs.Web.Addr)
		apiErr <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT)

	select {
	case err := <-apiErr:
		return err
	case sig := <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), envs.Web.ShutdownTimeout)
		defer cancel()

		log.Info().Msg("main: Shutdown server")
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrapf(err, "failed to shutdown server")
		}
		log.Info().Msgf("main: Terminated via signal %q", sig)
	}

	return nil
}
