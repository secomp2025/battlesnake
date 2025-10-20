package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/game"
	"github.com/secomp2025/localsnake/models"
	"github.com/secomp2025/localsnake/templates/pages"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	team_code := GetCookieValue(r, "team_code")
	if team_code == "" {
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return
	}

	codes := controllers.NewCodeController(database.DB)
	teams := controllers.NewTeamController(database.DB)
	snakes := controllers.NewSnakeController(database.DB)

	code, err := codes.FindCode(r.Context(), team_code)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if code.Code != "ADMBSNAKE" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	teamsList, err := teams.ListTeams(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var modelTeamList []models.Team
	for _, team := range teamsList {

		if team.IsAdmin.Valid && team.IsAdmin.Bool {
			continue
		}

		code, err := codes.GetCodeByTeam(r.Context(), team.ID)
		if err != nil || code == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		modelTeamList = append(modelTeamList, models.Team{
			ID:   team.ID,
			Name: team.Name,
			Code: code.Code,
		})
	}

	for i, team := range modelTeamList {
		teamSnakes, err := snakes.ListTeamSnakes(r.Context(), team.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// get the last snake
		if len(teamSnakes) == 0 {
			continue
		}

		lastSnake := teamSnakes[len(teamSnakes)-1]

		status, err := controllers.GetServerManager().GetSnakeStatus(r.Context(), &lastSnake)
		if err != nil {
			status = game.StatusOffline
		}

		modelTeamList[i].Snake = &models.Snake{
			ID:        lastSnake.ID,
			UpdatedAt: lastSnake.UpdatedAt.Time,
			Lang:      lastSnake.Lang,
			Status:    string(status),
		}
	}

    // build codes list with claimed status
    codeList, err := codes.ListCodes(r.Context())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    var modelCodes []models.Code
    for _, c := range codeList {
        t, err := teams.GetTeamByCode(r.Context(), c.ID)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        used := t != nil
        modelCodes = append(modelCodes, models.Code{ID: c.ID, Code: c.Code, Used: used})
    }

	templ.Handler(pages.Admin(modelTeamList, modelCodes)).ServeHTTP(w, r)
}
