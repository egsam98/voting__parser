package rest

import (
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"

	"github.com/egsam98/voting/parser/services/votes"
)

func API(votes *votes.Service, saramaClient sarama.Client) (http.Handler, error) {
	mux := chi.NewMux()
	mux.Use(
		middleware.Recoverer,
		middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger: &log.Logger,
		}),
	)

	vc := newVoteController(votes)

	hc, err := newHealthController(saramaClient)
	if err != nil {
		return nil, err
	}

	mux.Post("/vote", vc.Handle)
	mux.Route("/health", func(r chi.Router) {
		r.Get("/readiness", hc.Readiness)
	})

	return mux, nil
}
