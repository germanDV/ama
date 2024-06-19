package questionnaire

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/germandv/ama/internal/uid"
)

type IService interface {
	Create(title string) (Questionnaire, error)
	Ask(questionnaireID string, text string) (Question, error)
	Get(questionnaireID string) ([]Question, error)
	GetMeta(questionnaireID string) (Questionnaire, error)
	CountQuestionnaires() (int, error)
	CountQuestions(questionnaireID string) (int, error)
	DeleteQuestionnaire(questionnaireID string) error
	Vote(questionnaireID string, questionID string, voterID string) (uint16, error)
	Answer(questionnaireID string, questionID string) error
}

type ballot struct {
	id         string
	expiration time.Time
	voters     map[string]bool
}

type Service struct {
	repo    Repository
	mu      sync.Mutex
	ttl     time.Duration
	logger  *slog.Logger
	ballots map[string]ballot
}

func NewService(repo Repository, ttl time.Duration, logger *slog.Logger) IService {
	svc := &Service{
		repo:    repo,
		mu:      sync.Mutex{},
		ttl:     ttl,
		logger:  logger,
		ballots: make(map[string]ballot),
	}

	go svc.gcExpiredBallots()
	return svc
}

func (s *Service) newBallot(id string) ballot {
	return ballot{
		id:         id,
		expiration: time.Now().Add(s.ttl),
		voters:     make(map[string]bool),
	}
}

func (s *Service) gcExpiredBallots() {
	for range time.Tick(3 * time.Second) {
		now := time.Now()
		s.logger.Debug("started gcExpiredBallots", "ballots", len(s.ballots))
		for id, ballot := range s.ballots {
			if now.After(ballot.expiration) {
				delete(s.ballots, id)
			}
		}
		s.logger.Debug("finished gcExpiredBallots", "ballots", len(s.ballots))
	}
}

func (s *Service) Vote(questionnaireID string, questionID string, voterID string) (uint16, error) {
	s.mu.Lock()

	_, found := s.ballots[questionID]
	if !found {
		s.ballots[questionID] = s.newBallot(questionID)
	}

	hasVoted, found := s.ballots[questionID].voters[voterID]
	if !found || !hasVoted {
		s.ballots[questionID].voters[voterID] = true
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		return 0, fmt.Errorf("%s already voted %s", voterID, questionID)
	}

	return s.repo.Vote(questionnaireID, questionID)
}

func (s *Service) Create(title string) (Questionnaire, error) {
	q := NewQuestionnaire(title)
	err := s.repo.SaveQuestionnaire(q)
	if err != nil {
		return Questionnaire{}, err
	}
	return q, nil
}

func (s *Service) Ask(questionnaireID string, text string) (Question, error) {
	q := NewQuestion(uid.Generate(false, 16), questionnaireID, text)
	err := s.repo.SaveQuestion(questionnaireID, q)
	if err != nil {
		return Question{}, err
	}
	return q, nil
}

func (s *Service) Get(questionnaireID string) ([]Question, error) {
	return s.repo.GetQuestions(questionnaireID)
}

func (s *Service) GetMeta(questionnaireID string) (Questionnaire, error) {
	return s.repo.GetQuestionnaire(questionnaireID)
}

func (s *Service) Answer(questionnaireID string, questionID string) error {
	return s.repo.Answer(questionnaireID, questionID)
}

func (s *Service) CountQuestionnaires() (int, error) {
	return s.repo.CountQuestionnaires()
}

func (s *Service) CountQuestions(questionnaireID string) (int, error) {
	return s.repo.CountQuestions(questionnaireID)
}

func (s *Service) DeleteQuestionnaire(questionnaireID string) error {
	return s.repo.DeleteQuestionnaire(questionnaireID)
}
