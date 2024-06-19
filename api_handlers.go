package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/webutils"
	"github.com/germandv/ama/internal/wsmanager"
)

func newQuestionnaireHandler(svc questionnaire.IService, web webutils.Web) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Req struct {
			Title string `json:"title"`
		}

		req := &Req{}
		ok := web.DecodeBody(w, r, req)
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

func newQuestionHandler(
	svc questionnaire.IService,
	wsm *wsmanager.WSManager,
	web webutils.Web,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Req struct {
			Question string `json:"question"`
		}

		req := &Req{}
		ok := web.DecodeBody(w, r, req)
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

func getQuestionsHandler(svc questionnaire.IService, web webutils.Web) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			web.BadRequest(w, errors.New("no questionnaire ID provided"))
			return
		}

		qs, err := svc.Get(questionnaireID)
		if err != nil {
			web.NotFound(w, "questionnaire", questionnaireID)
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

func voteHandler(
	svc questionnaire.IService,
	wsm *wsmanager.WSManager,
	web webutils.Web,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			web.BadRequest(w, errors.New("no questionnaire ID provided"))
			return
		}

		questionID := r.PathValue("question_id")
		if questionID == "" {
			web.BadRequest(w, errors.New("no question ID provided"))
			return
		}

		cookie, err := r.Cookie("voter")
		if err != nil {
			web.Forbidden(w)
			return
		}

		count, err := svc.Vote(questionnaireID, questionID, cookie.Value)
		if err != nil {
			web.BadRequest(w, err)
			return
		}

		msg := newVoteMessage(questionID, count)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			web.InternalError(w, err)
			return
		}
		wsm.Broadcast(questionnaireID, jsonMsg)

		w.WriteHeader(http.StatusOK)
	}
}

func answerHandler(
	svc questionnaire.IService,
	wsm *wsmanager.WSManager,
	web webutils.Web,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			web.BadRequest(w, errors.New("no questionnaire ID provided"))
			return
		}

		questionID := r.PathValue("question_id")
		if questionID == "" {
			web.BadRequest(w, errors.New("no question ID provided"))
			return
		}

		meta, err := svc.GetMeta(questionnaireID)
		if err != nil {
			web.NotFound(w, "questionnaire", questionnaireID)
			return
		}

		cookie, err := r.Cookie("host")
		if err != nil || cookie.Value != meta.Host {
			web.Forbidden(w)
			return
		}

		err = svc.Answer(questionnaireID, questionID)
		if err != nil {
			web.InternalError(w, err)
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
