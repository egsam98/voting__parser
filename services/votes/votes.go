package votes

import (
	"github.com/Shopify/sarama"
	votingpb "github.com/egsam98/voting/proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	prod           sarama.SyncProducer
	validatorTopic string
}

func NewService(prod sarama.SyncProducer, validatorTopic string) *Service {
	return &Service{prod: prod, validatorTopic: validatorTopic}
}

func (s *Service) RequestValidation(candidateID int64, passport, fullname string) error {
	vote := &votingpb.Vote{
		Voter: &votingpb.Voter{
			Passport: passport,
			Fullname: fullname,
		},
		CandidateId: candidateID,
	}

	b, err := proto.Marshal(vote)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal vote %#v", vote)
	}

	_, _, err = s.prod.SendMessage(&sarama.ProducerMessage{
		Topic: s.validatorTopic,
		Value: sarama.ByteEncoder(b),
	})
	return errors.Wrapf(err, "failed to send %#v to topic=%s", vote, s.validatorTopic)
}
