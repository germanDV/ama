package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/wsmanager"
)

func homePageHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("home.html"))
	tmpl.Execute(w, map[string]any{"Server": "http://localhost:3000"})
}

func questionnairePageHandler(svc questionnaire.IService) http.HandlerFunc {
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
			"Server":    "http://localhost:3000",
			"ServerWS":  "ws://localhost:3000/ws?questionnaire=" + meta.ID,
			"ID":        meta.ID,
			"Title":     meta.Title,
			"Questions": qs,
			"IsHost":    isHost,
		}

		tmpl.Execute(w, data)
	}
}

func main() {
	wsm := wsmanager.New()
	svc := questionnaire.NewService(questionnaire.NewInMemoryRepo())

	qLimiter := globalLimiter(20, svc.CountQuestionnaires)
	qsLimiter := idLimiter(100, svc.CountQuestions)
	cLimiter := idLimiter(100, wsm.CountClients)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", wsHandler(wsm, svc))
	mux.HandleFunc("GET /", homePageHandler)
	mux.Handle("GET /{id}", cLimiter(questionnairePageHandler(svc)))
	mux.Handle("POST /questionnaires", qLimiter(newQuestionnaireHandler(svc)))
	mux.Handle("POST /questionnaires/{id}/questions", qsLimiter(newQuestionHandler(svc, wsm)))
	mux.HandleFunc("GET /questionnaires/{id}/questions", getQuestionsHandler(svc))
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/vote", voteHandler(svc, wsm))
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/answer", answerHandler(svc, wsm))

	server := &http.Server{
		Addr:              ":3000",
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Println("Connect to WS on ws://localhost:3000/ws")
	log.Println("Use client on http://localhost:3000 (browser)")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
