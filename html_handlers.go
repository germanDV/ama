package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
)

func homePageHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		tmpl := template.Must(template.ParseFiles("home.html"))
		url := fmt.Sprintf("http://localhost:%d", port)
		tmpl.Execute(w, map[string]any{"Server": url})
	}
}

func questionnairePageHandler(svc questionnaire.IService, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("q.html"))

		questionnaireID := r.PathValue("id")
		if questionnaireID == "" {
			http.Error(w, "no questionnaire ID provided", http.StatusBadRequest)
			return
		}

		meta, err := svc.GetMeta(questionnaireID)
		if err != nil {
			http.Error(w, "no questionnaire found", http.StatusNotFound)
			return
		}

		qs, err := svc.Get(questionnaireID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error fetching existing questions", http.StatusInternalServerError)
			return
		}

		isHost := false
		cookie, err := r.Cookie("host")
		if err == nil && meta.Host == cookie.Value {
			isHost = true
		}

		data := map[string]any{
			"Server":    fmt.Sprintf("http://localhost:%d", port),
			"ServerWS":  fmt.Sprintf("ws://localhost:%d/ws?questionnaire=%s", port, meta.ID),
			"ID":        meta.ID,
			"Title":     meta.Title,
			"Questions": qs,
			"IsHost":    isHost,
		}

		tmpl.Execute(w, data)
	}
}
