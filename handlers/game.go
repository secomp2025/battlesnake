package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/game"
)

// type boardServerHandler struct {
// 	game   board.Game
// 	events chan board.GameEvent
// 	done   chan bool
// }

type globalBoardServerHandler struct {
	lock     sync.Mutex
	handlers map[string]*game.BoardServer
}

var boardManager globalBoardServerHandler

func init() {
	boardManager = globalBoardServerHandler{
		lock:     sync.Mutex{},
		handlers: make(map[string]*game.BoardServer),
	}
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	team_code := GetCookieValue(r, "team_code")
	if team_code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Println("creating game for team code: ", team_code)

	codes := controllers.NewCodeController(database.DB)
	code, err := codes.FindCode(r.Context(), team_code)
	if err != nil || code == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teams := controllers.NewTeamController(database.DB)
	team, err := teams.GetTeamByCode(r.Context(), code.ID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	snakes := controllers.NewSnakeController(database.DB)
	team_snakes, err := snakes.ListTeamSnakes(r.Context(), team.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var gameSnakes []game.Snake

	for _, snake := range team_snakes {
		snakeServer := controllers.GetServerManager().GetServer(snake.ID)
		if snakeServer == nil {
			continue
		}
		gameSnakes = append(gameSnakes, game.Snake{
			Name: team.Name,
			URL:  snakeServer.Addr,
		})
	}

	if len(gameSnakes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// games := controllers.NewGameController()
	gameController := controllers.NewGameController()
	gameInfo := gameController.CreateGame(gameSnakes)

	log.Println("game id: ", gameInfo.ID)

	boardManager.lock.Lock()
	defer boardManager.lock.Unlock()

	boardManager.handlers[gameInfo.ID] = gameInfo.Server

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, gameInfo.ID)
}

// Handle /game/<game_id> and /game/<game_id>/events
func GameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Println("handling game id: ", r.URL.Path)

	splits := strings.Split(r.URL.Path, "/")
	if len(splits) < 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	gameID := splits[2]
	is_event := len(splits) > 3 && splits[3] == "events"

	boardManager.lock.Lock()
	defer boardManager.lock.Unlock()

	if game, ok := boardManager.handlers[gameID]; ok {
		if is_event {
			game.HandleGame(w, r)
			return
		} else {
			if r.Header.Get("Connection") != "Upgrade" || r.Header.Get("Upgrade") != "websocket" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			game.HandleWebsocket(w, r)
		}

		// already have a handler for this game
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
