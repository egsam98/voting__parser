package requests

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

var _ render.Binder = (*Vote)(nil)

type Vote struct {
	CandidateID int64 `json:"candidate_id"`
	Voter       struct {
		Passport string `json:"passport"`
		Fullname string `json:"fullname"`
	} `json:"voter"`
}

func (v *Vote) Bind(*http.Request) error {
	if v.CandidateID == 0 {
		return errors.New("\"candidate_id\" must be non empty")
	}
	if v.Voter.Passport == "" {
		return errors.New("\"voter.passport\" must be non empty")
	}
	if v.Voter.Fullname == "" {
		return errors.New("\"voter.fullname\" must be non empty")
	}
	return nil
}
