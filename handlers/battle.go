package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/game"
	"github.com/secomp2025/localsnake/templates/pages"
)

func HandleBattle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snake_ids_param := r.URL.Query().Get("snake_ids")
	if snake_ids_param == "" {
		http.Error(w, "snake_ids is required", http.StatusBadRequest)
		return
	}

	snake_ids := strings.Split(snake_ids_param, ",")
	if len(snake_ids) == 0 {
		http.Error(w, "snake_ids must be a comma separated list of two snake ids", http.StatusBadRequest)
		return
	}

	snakes := controllers.NewSnakeController(database.DB)
	teams := controllers.NewTeamController(database.DB)

	var gameSnakes []game.Snake

	for _, snakeID := range snake_ids {
		snakeIDInt, err := strconv.ParseInt(snakeID, 10, 64)
		if err != nil {
			http.Error(w, "snake id must be an integer", http.StatusBadRequest)
			return
		}
		snake, err := snakes.GetSnake(r.Context(), snakeIDInt)
		if err != nil {
			http.Error(w, "snake not found", http.StatusNotFound)
			return
		}
		if snake == nil {
			log.Println("snake not found:", snakeID)
			http.Error(w, "snake not found", http.StatusNotFound)
			return
		}

		snakeTeam, err := teams.GetTeam(r.Context(), snake.TeamID)
		if err != nil {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		if snakeTeam == nil {
			log.Println("team not found:", snake.TeamID)
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}

		snakeServer := controllers.GetServerManager().GetServer(snake.ID)
		if snakeServer == nil {
			log.Println("snake server not found:", snake.ID)

			err := controllers.GetServerManager().ManageSnake(r.Context(), snake)
			if err != nil {
				log.Println("snake server not created:", snake.ID)
				http.Error(w, "snake server not created", http.StatusInternalServerError)
				return
			}

			snakeServer = controllers.GetServerManager().GetServer(snake.ID)
			if snakeServer == nil {
				log.Println("snake server not found after creation:", snake.ID)
				http.Error(w, "snake server not found", http.StatusInternalServerError)
				return
			}
		}

		gameSnake := game.Snake{
			Name: snakeTeam.Name,
			URL:  snakeServer.Addr,
		}
		gameSnakes = append(gameSnakes, gameSnake)
	}

	if len(gameSnakes) == 0 {
		http.Error(w, "no snakes found", http.StatusNotFound)
		return
	}

	log.Println("game snakes: ", len(gameSnakes))

	gameController := controllers.NewGameController()
	gameInfo := gameController.CreateGame(gameSnakes)

	log.Println("game id: ", gameInfo.ID)

	boardManager.lock.Lock()
	defer boardManager.lock.Unlock()

	boardManager.handlers[gameInfo.ID] = gameInfo.Server

	templ.Handler(pages.Battle(gameInfo.ID)).ServeHTTP(w, r)
}
