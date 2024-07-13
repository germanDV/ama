package questionnaire

type Metadata struct {
	Votes    uint16 `json:"votes"`
	Answered bool   `json:"answered"`
}

type Question struct {
	ID            string   `json:"id"`
	Questionnaire string   `json:"questionnaire"`
	Question      string   `json:"question"`
	Metadata      Metadata `json:"metadata"`
}

func NewQuestion(id string, questionnaire string, question string) Question {
	return Question{
		ID:            id,
		Questionnaire: questionnaire,
		Question:      question,
		Metadata: Metadata{
			Votes:    0,
			Answered: false,
		},
	}
}
