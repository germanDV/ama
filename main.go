package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/wsmanager"
)

func main() {
	port := 3000
	wsm := wsmanager.New()
	svc := questionnaire.NewService(questionnaire.NewInMemoryRepo())

	qLimiter := globalLimiter(20, svc.CountQuestionnaires)
	qsLimiter := idLimiter(100, svc.CountQuestions)
	cLimiter := idLimiter(100, wsm.CountClients)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", wsHandler(wsm, svc))
	mux.HandleFunc("GET /", homePageHandler(port))
	mux.Handle("GET /{id}", cLimiter(questionnairePageHandler(svc, port)))
	mux.Handle("POST /questionnaires", qLimiter(newQuestionnaireHandler(svc)))
	mux.Handle("POST /questionnaires/{id}/questions", qsLimiter(newQuestionHandler(svc, wsm)))
	mux.HandleFunc("GET /questionnaires/{id}/questions", getQuestionsHandler(svc))
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/vote", voteHandler(svc, wsm))
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/answer", answerHandler(svc, wsm))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	// TODO: add graceful shutdown
	log.Printf("Connect to WS on ws://localhost:%d/ws\n", port)
	log.Printf("Use client on http://localhost:%d (browser)\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
