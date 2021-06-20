package rest

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"

	"github.com/egsam98/voting/parser/services/votes"
)

func API(votes *votes.Service) http.Handler {
	mux := chi.NewMux()
	mux.Use(
		middleware.Recoverer,
		middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger: &log.Logger,
		}),
	)

	vc := newVoteController(votes)

	mux.Post("/vote", vc.Handle)

	return mux
}
