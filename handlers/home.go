package handlers

import (
	"database/sql"
	"net/http"

	"github.com/a-h/templ"
	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/templates/pages"
)

// HomePage renders the dashboard if the user has a valid team_code cookie;
// otherwise redirects to /login. Validation is mocked via mockBound.
func HomePage(w http.ResponseWriter, r *http.Request) {
	team_code := GetCookieValue(r, "team_code")
	if team_code == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	codes := controllers.NewCodeController(database.DB)

	code, err := codes.FindCode(r.Context(), team_code)
	if err != nil || code == nil {
		ClearCookie(w, "team_code")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	teams := controllers.NewTeamController(database.DB)
	team, err := teams.GetTeamByCode(r.Context(), code.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/login?code="+team_code, http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// For now we mock the status; ensure it conforms to fixed values
    // Pass admin flag for conditional Admin nav on dashboard
    isAdmin := team.IsAdmin.Valid && team.IsAdmin.Bool
    templ.Handler(pages.Dashboard(team.Name, isAdmin)).ServeHTTP(w, r)
}
