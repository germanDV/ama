package main

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/uid"
	"github.com/germandv/ama/internal/webutils"
)

func homePageHandler(web webutils.Web) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		tmpl := template.Must(template.ParseFiles("views/layout.html", "views/home.html"))
		tmpl.Execute(w, map[string]any{"Server": web.GetApiURL()})
	}
}

func questionnairePageHandler(svc questionnaire.IService, web webutils.Web) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("views/layout.html", "views/q.html"))

		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			web.BadRequest(w, errors.New("no questionnaire ID provided"))
			return
		}

		meta, err := svc.GetMeta(questionnaireID)
		if err != nil {
			web.NotFound(w, "questionnaire", questionnaireID)
			return
		}

		qs, err := svc.Get(questionnaireID)
		if err != nil {
			web.InternalError(w, errors.New("error fetching existing questions"))
			return
		}

		isHost := false
		cookie, err := r.Cookie("host")
		if err == nil && meta.Host == cookie.Value {
			isHost = true
		}

		hasVoterCookie := false
		_, err = r.Cookie("voter")
		if err == nil {
			hasVoterCookie = true
		}

		if !isHost && !hasVoterCookie {
			voterID := uid.Generate(false, 32)
			web.SetCookie(w, "voter", voterID)
		}

		data := map[string]any{
			"Server":    web.GetApiURL(),
			"ServerWS":  web.GetWebsocketURL(meta.ID),
			"ID":        meta.ID,
			"Title":     meta.Title,
			"Questions": qs,
			"IsHost":    isHost,
		}

		tmpl.Execute(w, data)
	}
}
