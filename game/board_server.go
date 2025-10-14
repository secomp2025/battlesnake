package game

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BattlesnakeOfficial/rules/board"
	"github.com/gorilla/websocket"
	log "github.com/spf13/jwalterweatherman"
)

// A minimal server capable of handling the requests from a single browser client running the board viewer.
type BoardServer struct {
	game   board.Game
	events chan board.GameEvent // channel for sending events from the game runner to the browser client
	done   chan bool            // channel for signalling (via closing) that all events have been sent to the browser client
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewBoardServer(game board.Game) *BoardServer {

	server := &BoardServer{
		game:   game,
		events: make(chan board.GameEvent, 1000), // buffered channel to allow game to run ahead of browser client
		done:   make(chan bool),
	}

	return server
}

// Handle the /games/:id request made by the board to fetch the game metadata.
func (server *BoardServer) HandleGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Game board.Game
	}{server.game})
	if err != nil {
		log.ERROR.Printf("Unable to serialize game: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Handle the /games/:id/events websocket request made by the board to receive game events.
func (server *BoardServer) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.ERROR.Printf("Unable to upgrade connection: %v", err)
		return
	}

	defer func() {
		err = ws.Close()
		if err != nil {
			log.ERROR.Printf("Unable to close websocket stream")
		}
	}()

	for event := range server.events {
		jsonStr, err := json.Marshal(event)
		if err != nil {
			log.ERROR.Printf("Unable to serialize event for websocket: %v", err)
		}

		err = ws.WriteMessage(websocket.TextMessage, jsonStr)
		if err != nil {
			log.ERROR.Printf("Unable to write to websocket: %v", err)
			break
		}
	}

	log.DEBUG.Printf("Finished writing all game events, signalling game server to stop")
	close(server.done)

	log.DEBUG.Printf("Sending websocket close message")
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.ERROR.Printf("Problem closing websocket: %v", err)
	}
}

func (server *BoardServer) Shutdown() {
	close(server.events)

	// wait for at max 10 seconds to allow clients to finish
	select {
	case <-server.done:
		log.DEBUG.Printf("Server is done, exiting")
	case <-time.After(10 * time.Second):
		log.DEBUG.Printf("Server timed out, exiting")
	}
}

func (server *BoardServer) SendEvent(event board.GameEvent) {
	server.events <- event
}
