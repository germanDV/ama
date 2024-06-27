package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/germandv/ama/internal/webutils"
	"github.com/germandv/ama/internal/wsmanager"
)

type MessageEvent string

const (
	MessageEventNewQuestion = MessageEvent("new_question")
	MessageEventVote        = MessageEvent("vote")
	MessageAnswer           = MessageEvent("answer")
)

type QuestionMessage struct {
	Event   MessageEvent `json:"event"`
	Details struct {
		ID       string `json:"id"`
		Question string `json:"question"`
		Votes    uint16 `json:"votes"`
	} `json:"details"`
}

func newQuestionMessage(id string, question string, votes uint16) QuestionMessage {
	return QuestionMessage{
		Event: MessageEventNewQuestion,
		Details: struct {
			ID       string `json:"id"`
			Question string `json:"question"`
			Votes    uint16 `json:"votes"`
		}{
			ID:       id,
			Question: question,
			Votes:    votes,
		},
	}
}

type VoteMessage struct {
	Event   MessageEvent `json:"event"`
	Details struct {
		ID    string `json:"id"`
		Votes uint16 `json:"votes"`
	} `json:"details"`
}

func newVoteMessage(id string, votes uint16) VoteMessage {
	return VoteMessage{
		Event: MessageEventVote,
		Details: struct {
			ID    string `json:"id"`
			Votes uint16 `json:"votes"`
		}{
			ID:    id,
			Votes: votes,
		},
	}
}

type AnswerMessage struct {
	Event   MessageEvent `json:"event"`
	Details struct {
		ID string `json:"id"`
	} `json:"details"`
}

func newAnswerMessage(id string) AnswerMessage {
	return AnswerMessage{
		Event: MessageAnswer,
		Details: struct {
			ID string `json:"id"`
		}{
			ID: id,
		},
	}
}

func wsHandler(
	wsm *wsmanager.WSManager,
	svc questionnaire.IService,
	logger *slog.Logger,
	web webutils.Web,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaire := r.URL.Query().Get("questionnaire")
		if questionnaire == "" {
			web.BadRequest(w, errors.New("no questionnaire ID provided"))
			return
		}

		c, err := wsm.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			web.InternalError(w, errors.New("error upgrading connection"))
			return
		}

		wsm.AddClient(questionnaire, c)
		wsm.Stats()

		defer func() {
			c.Close()
			wsm.RemoveClient(questionnaire, c)

			count, err := wsm.CountClients(questionnaire)
			if err == nil && count == 0 {
				wsm.DeleteRoom(questionnaire)
			}

			wsm.Stats()
		}()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				if wsm.IsCloseError(err) {
					logger.Debug("WS client disconnected")
					break
				}
				logger.Warn("error reading WS message, disconnecting", "err", err)
				break
			}
		}
	}
}
