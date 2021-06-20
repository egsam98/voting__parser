package rest

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/egsam98/voting/parser/handlers/rest/requests"
	"github.com/egsam98/voting/parser/internal/web"
	"github.com/egsam98/voting/parser/services/votes"
)

type voteController struct {
	votes *votes.Service
}

func newVoteController(votes *votes.Service) *voteController {
	return &voteController{votes: votes}
}

func (vc *voteController) Handle(w http.ResponseWriter, r *http.Request) {
	var req requests.Vote
	if err := render.Bind(r, &req); err != nil {
		web.RespondError(w, r, web.WrapWithError(votes.ErrInvalidInput, err))
		return
	}

	if err := vc.votes.RequestValidation(req.CandidateID, req.Voter.Passport, req.Voter.Fullname); err != nil {
		web.RespondError(w, r, err)
	}
}
