package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/webutils"
	"github.com/germandv/ama/internal/wsmanager"
)

func main() {
	domain := "localhost"
	port := 3000
	secure := false
	logLevel := slog.LevelDebug

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	web := webutils.New(24*time.Hour, logger, domain, port, secure)
	wsm := wsmanager.New()
	svc := questionnaire.NewService(questionnaire.NewInMemoryRepo())

	qLimiter := globalLimiter(20, svc.CountQuestionnaires, logger, web)
	qsLimiter := idLimiter(100, svc.CountQuestions, logger, web)
	cLimiter := idLimiter(100, wsm.CountClients, logger, web)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", wsHandler(wsm, svc, logger, web))
	mux.HandleFunc("GET /", homePageHandler(web))
	mux.Handle("GET /{id}", cLimiter(questionnairePageHandler(svc, web)))
	mux.Handle("POST /questionnaires", qLimiter(newQuestionnaireHandler(svc, web)))
	mux.Handle("POST /questionnaires/{id}/questions", qsLimiter(newQuestionHandler(svc, wsm, web)))
	mux.HandleFunc("GET /questionnaires/{id}/questions", getQuestionsHandler(svc, web))
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
	logger.Info("WS server up", "port", port)
	logger.Info("HTTP server up", "port", port)
	err := server.ListenAndServe()
	if err != nil {
		logger.Error("Server error", "err", err)
		panic(err)
	}
}
