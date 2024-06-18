package main

import (
	"encoding/json"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/web"
	"github.com/germandv/ama/internal/wsmanager"
)

func newQuestionnaireHandler(svc questionnaire.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Req struct {
			Title string `json:"title"`
		}

		req, ok := web.DecodeBody[Req](w, r)
		if !ok {
			return
		}

		q, err := svc.Create(req.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		web.SetCookie(w, "host", q.Host)
		web.JSON(w, http.StatusCreated, q)
	}
}

func newQuestionHandler(svc questionnaire.IService, wsm *wsmanager.WSManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Req struct {
			Question string `json:"question"`
		}

		req, ok := web.DecodeBody[Req](w, r)
		if !ok {
			return
		}

		questionnaire := r.PathValue("id")
		if questionnaire == "" {
			http.Error(w, "no questionnaire ID provided", http.StatusBadRequest)
			return
		}

		q, err := svc.Ask(questionnaire, req.Question)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg := newQuestionMessage(q.ID, q.Question, q.Metadata.Votes)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wsm.Broadcast(questionnaire, jsonMsg)

		web.JSON(w, http.StatusCreated, q)
	}
}

func getQuestionsHandler(svc questionnaire.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			http.Error(w, "no questionnaire ID provided", http.StatusBadRequest)
			return
		}

		qs, err := svc.Get(questionnaireID)
		if err != nil {
			http.Error(w, "no questionnaire found", http.StatusNotFound)
			return
		}

		envelope := struct {
			Questions []questionnaire.Question `json:"questions"`
		}{
			Questions: qs,
		}
		web.JSON(w, http.StatusOK, envelope)
	}
}

func voteHandler(svc questionnaire.IService, wsm *wsmanager.WSManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			http.Error(w, "no questionnaire ID provided", http.StatusBadRequest)
			return
		}

		questionID := r.PathValue("question_id")
		if questionID == "" {
			http.Error(w, "no questionID ID provided", http.StatusBadRequest)
			return
		}

		count, err := svc.Vote(questionnaireID, questionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		msg := newVoteMessage(questionID, count)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wsm.Broadcast(questionnaireID, jsonMsg)

		w.WriteHeader(http.StatusOK)
	}
}

func answerHandler(svc questionnaire.IService, wsm *wsmanager.WSManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			http.Error(w, "no questionnaire ID provided", http.StatusBadRequest)
			return
		}

		questionID := r.PathValue("question_id")
		if questionID == "" {
			http.Error(w, "no questionID ID provided", http.StatusBadRequest)
			return
		}

		meta, err := svc.GetMeta(questionnaireID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		cookie, err := r.Cookie("host")
		if err != nil || cookie.Value != meta.Host {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		err = svc.Answer(questionnaireID, questionID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		msg := newAnswerMessage(questionID)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wsm.Broadcast(questionnaireID, jsonMsg)

		w.WriteHeader(http.StatusOK)
	}
}
