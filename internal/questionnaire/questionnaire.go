package questionnaire

import "github.com/germandv/ama/internal/uid"

type Questionnaire struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Host  string `json:"host"`
}

func NewQuestionnaire(title string) Questionnaire {
	return Questionnaire{
		ID:    uid.Generate(true, 16),
		Title: title,
		Host:  uid.Generate(false, 32),
	}
}
