package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/webutils"
	"github.com/germandv/ama/internal/wsmanager"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	web := webutils.New(cfg.TTL, logger, cfg.Domain, cfg.Port, cfg.Secure)
	wsm := wsmanager.New()
	repo := questionnaire.NewRedisRepo(cfg.RedisHost, cfg.RedisPort, cfg.RedisPass, cfg.TTL)
	svc := questionnaire.NewService(repo, cfg.TTL, logger)

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
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/vote", voteHandler(svc, wsm, web))
	mux.HandleFunc("PUT /questionnaires/{id}/questions/{question_id}/answer", answerHandler(svc, wsm, web))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           realIP(mux),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	killCh := make(chan os.Signal, 1)
	signal.Notify(killCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", "err", err)
			os.Exit(1)
		}
	}()

	logger.Info("WS server up", "port", cfg.Port)
	logger.Info("HTTP server up", "port", cfg.Port)

	<-killCh
	logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error("Failed to shut down gracefully", "err", err)
		cancel()
		os.Exit(1)
	}

	cancel()
	logger.Info("Shutdown completed")
}
