package questionnaire

type IService interface {
	Create(title string) (Questionnaire, error)
	Ask(questionnaireID string, text string) (Question, error)
	Get(questionnaireID string) ([]Question, error)
	GetMeta(questionnaireID string) (Questionnaire, error)
	CountQuestionnaires() (int, error)
	CountQuestions(questionnaireID string) (int, error)
	DeleteQuestionnaire(questionnaireID string) error
	Vote(questionnaireID string, questionID string) (uint16, error)
	Answer(questionnaireID string, questionID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) IService {
	return Service{repo}
}

func (s Service) Create(title string) (Questionnaire, error) {
	q := NewQuestionnaire(title)
	err := s.repo.SaveQuestionnaire(q)
	if err != nil {
		return Questionnaire{}, err
	}
	return q, nil
}

func (s Service) Ask(questionnaireID string, text string) (Question, error) {
	q := NewQuestion(generateID(false, 16), questionnaireID, text)
	err := s.repo.SaveQuestion(questionnaireID, q)
	if err != nil {
		return Question{}, err
	}
	return q, nil
}

func (s Service) Get(questionnaireID string) ([]Question, error) {
	return s.repo.GetQuestions(questionnaireID)
}

func (s Service) GetMeta(questionnaireID string) (Questionnaire, error) {
	return s.repo.GetQuestionnaire(questionnaireID)
}

func (s Service) Vote(questionnaireID string, questionID string) (uint16, error) {
	return s.repo.Vote(questionnaireID, questionID)
}

func (s Service) Answer(questionnaireID string, questionID string) error {
	return s.repo.Answer(questionnaireID, questionID)
}

func (s Service) CountQuestionnaires() (int, error) {
	return s.repo.CountQuestionnaires()
}

func (s Service) CountQuestions(questionnaireID string) (int, error) {
	return s.repo.CountQuestions(questionnaireID)
}

func (s Service) DeleteQuestionnaire(questionnaireID string) error {
	return s.repo.DeleteQuestionnaire(questionnaireID)
}
