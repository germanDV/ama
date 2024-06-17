package wsmanager

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

// WSManager handles a WebSocket, its rooms and their clients.
type WSManager struct {
	rooms    map[string]map[*websocket.Conn]bool
	Upgrader websocket.Upgrader
}

// New creates a WSManager.
func New() *WSManager {
	return &WSManager{
		rooms:    make(map[string]map[*websocket.Conn]bool),
		Upgrader: websocket.Upgrader{},
	}
}

// AddClient adds a client connection to a room.
// If the room does not exists, it creates it.
func (wsm *WSManager) AddClient(id string, c *websocket.Conn) {
	clients, found := wsm.rooms[id]
	if !found {
		clients = make(map[*websocket.Conn]bool)
		wsm.rooms[id] = clients
	}
	clients[c] = true
}

// RemoveClient removes a client connection from a room.
// It does nothing if the room or client do not exist.
func (wsm *WSManager) RemoveClient(id string, c *websocket.Conn) {
	clients, found := wsm.rooms[id]
	if !found {
		return
	}
	delete(clients, c)
}

// DeleteRoom deletes the room and all its clients connections (if any).
// It does nothing if the room does not exist.
func (wsm *WSManager) DeleteRoom(id string) {
	delete(wsm.rooms, id)
}

// Broadcast sends a message to all clients in the room (including emitter).
func (wsm *WSManager) Broadcast(room string, data []byte) error {
	clients, found := wsm.rooms[room]
	if !found {
		return fmt.Errorf("room %q does not exist", room)
	}
	for c := range clients {
		c.WriteMessage(1, data)
	}
	return nil
}

// CountClients counts clients connected to a room.
// If room does not exists, it returns 0.
func (wsm *WSManager) CountClients(room string) (int, error) {
	clients, found := wsm.rooms[room]
	if !found {
		return 0, nil
	}
	return len(clients), nil
}

// Stats returns the count of clients per room.
func (wsm *WSManager) Stats() string {
	sb := strings.Builder{}
	for room, clients := range wsm.rooms {
		sb.WriteString("room ")
		sb.WriteString(room)
		sb.WriteString(" has ")
		sb.WriteString(fmt.Sprintf("%d", len(clients)))
		sb.WriteString(" clients\n")
	}
	return sb.String()
}

// IsCloseError retruns true if error is an expected disconnection error.
func (wsm *WSManager) IsCloseError(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}
