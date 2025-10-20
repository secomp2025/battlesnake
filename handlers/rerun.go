package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
)

func RerunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	teamCode := GetCookieValue(r, "team_code")
	if teamCode == "" {
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return
	}

	codes := controllers.NewCodeController(database.DB)
	snakes := controllers.NewSnakeController(database.DB)

	code, err := codes.FindCode(r.Context(), teamCode)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if code.Code != "ADMBSNAKE" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get data from request { "snake_id": "1" } json
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	snakeID := data["snake_id"]
	if snakeID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	snakeIDFloat, ok := snakeID.(float64)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	snakeModel, err := snakes.GetSnake(r.Context(), int64(snakeIDFloat))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if snakeModel == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	controllers.GetServerManager().StopAndRemoveSnake(snakeModel.ID)
	if err := controllers.GetServerManager().ManageSnake(r.Context(), snakeModel); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
