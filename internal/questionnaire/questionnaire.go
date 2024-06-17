package questionnaire

type Questionnaire struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Host  string `json:"-"`
}

func NewQuestionnaire(title string) Questionnaire {
	return Questionnaire{
		ID:    generateID(true, 16),
		Title: title,
		Host:  generateID(false, 32),
	}
}
