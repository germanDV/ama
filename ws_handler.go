package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/germandv/ama/internal/questionnaire"
	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

var rooms = make(map[string]map[*websocket.Conn]bool)

func printStats() {
	for room, clients := range rooms {
		log.Printf("room %q has %d clients", room, len(clients))
	}
}

func broadcast(room string, data []byte) {
	clients, found := rooms[room]
	if !found {
		log.Printf("room %q does not exist", room)
		return
	}
	for c := range clients {
		c.WriteMessage(1, data)
	}
}

func wsHandler(svc questionnaire.IService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionnaire := r.URL.Query().Get("questionnaire")
		if questionnaire == "" {
			log.Print("no questionnaire ID provided")
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("error upgrading connection:", err)
			return
		}

		clients, found := rooms[questionnaire]
		if !found {
			clients = make(map[*websocket.Conn]bool)
			rooms[questionnaire] = clients
		}
		clients[c] = true
		printStats()

		defer func() {
			c.Close()
			delete(clients, c)
			if len(rooms[questionnaire]) == 0 {
				delete(rooms, questionnaire)
				err := svc.DeleteQuestionnaire(questionnaire)
				if err != nil {
					fmt.Printf("error deleting questionnaire: %s\n", err)
				}
			}
			printStats()
		}()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived,
				) {
					log.Println("disconnection")
					break
				}
				log.Println("error reading message, disconnecting:", err)
				break
			}
		}
	}
}
