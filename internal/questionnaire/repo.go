package questionnaire

import (
	"fmt"
	"sync"
)

type Repository interface {
	SaveQuestionnaire(q Questionnaire) error
	SaveQuestion(questionnaireID string, q Question) error
	GetQuestions(questionnaireID string) ([]Question, error)
	GetQuestionnaire(questionnaireID string) (Questionnaire, error)
	CountQuestionnaires() (int, error)
	CountQuestions(questionnaireID string) (int, error)
	Vote(questionnaireID string, questionID string) (uint16, error)
	Answer(questionnaireID string, questionID string) error
}

// TODO: create a repository using Redis and put a TTL on the questionnaires and questions.

type InMemoryRepository struct {
	mu             sync.RWMutex
	questionnaires map[string]Questionnaire
	questions      map[string][]Question
}

func NewInMemoryRepo() Repository {
	return &InMemoryRepository{
		questionnaires: make(map[string]Questionnaire),
		questions:      make(map[string][]Question),
	}
}

func (r *InMemoryRepository) SaveQuestionnaire(q Questionnaire) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.questionnaires[q.ID] = q
	r.questions[q.ID] = make([]Question, 0, 10)
	return nil
}

func (r *InMemoryRepository) SaveQuestion(questionnaireID string, q Question) error {
	_, found := r.questionnaires[questionnaireID]
	if !found {
		return fmt.Errorf("questionnaire %s not found", questionnaireID)
	}

	r.mu.Lock()
	r.questions[questionnaireID] = append(r.questions[questionnaireID], q)
	r.mu.Unlock()

	return nil
}

func (r *InMemoryRepository) GetQuestions(questionnaireID string) ([]Question, error) {
	qs, found := r.questions[questionnaireID]
	if !found {
		return nil, fmt.Errorf("questionnaire %s not found", questionnaireID)
	}
	return qs, nil
}

func (r *InMemoryRepository) GetQuestionnaire(questionnaireID string) (Questionnaire, error) {
	q, found := r.questionnaires[questionnaireID]
	if !found {
		return Questionnaire{}, fmt.Errorf("questionnaire %s not found", questionnaireID)
	}
	return q, nil
}

func (r *InMemoryRepository) Vote(questionnaireID string, questionID string) (uint16, error) {
	_, found := r.questions[questionnaireID]
	if !found {
		return 0, fmt.Errorf("questionnaire %s not found", questionnaireID)
	}

	found = false
	count := uint16(0)

	r.mu.Lock()
	for i, q := range r.questions[questionnaireID] {
		if q.ID == questionID && !q.Metadata.Answered {
			found = true
			r.questions[questionnaireID][i].Metadata.Votes++
			count = r.questions[questionnaireID][i].Metadata.Votes
		}
	}
	r.mu.Unlock()

	if !found {
		return 0, fmt.Errorf("question %s not found", questionID)
	}
	return count, nil
}

func (r *InMemoryRepository) Answer(questionnaireID string, questionID string) error {
	_, found := r.questions[questionnaireID]
	if !found {
		return fmt.Errorf("questionnaire %s not found", questionnaireID)
	}

	found = false
	r.mu.Lock()
	for i, q := range r.questions[questionnaireID] {
		if q.ID == questionID {
			found = true
			r.questions[questionnaireID][i].Metadata.Answered = true
		}
	}
	r.mu.Unlock()

	if !found {
		return fmt.Errorf("question %s not found", questionID)
	}
	return nil
}

func (r *InMemoryRepository) CountQuestionnaires() (int, error) {
	return len(r.questionnaires), nil
}

func (r *InMemoryRepository) CountQuestions(questionnaireID string) (int, error) {
	return len(r.questions[questionnaireID]), nil
}
