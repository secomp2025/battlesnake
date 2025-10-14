package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/game"
	"github.com/secomp2025/localsnake/templates/components"
)

// StatusHandler returns the current snake status badge fragment.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	team_code := GetCookieValue(r, "team_code")
	if team_code == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	codes := controllers.NewCodeController(database.DB)
	code, err := codes.FindCode(r.Context(), team_code)
	if err != nil || code == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teams := controllers.NewTeamController(database.DB)
	team, err := teams.GetTeamByCode(r.Context(), code.ID)
	if err != nil || team == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	teamSnakes := controllers.NewSnakeController(database.DB)
	team_snakes, err := teamSnakes.ListTeamSnakes(r.Context(), team.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(team_snakes) == 0 {
		templ.Handler(components.StatusBadge(string(game.StatusOffline))).ServeHTTP(w, r)
		return
	}

	latestSnake := team_snakes[len(team_snakes)-1]
	status, err := controllers.GetServerManager().GetSnakeStatus(r.Context(), &latestSnake)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templ.Handler(components.StatusBadge(string(status))).ServeHTTP(w, r)
}
