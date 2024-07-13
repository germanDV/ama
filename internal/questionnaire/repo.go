package questionnaire

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
