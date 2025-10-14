package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// isPrintableASCII returns true if all bytes are ASCII printable (space ' ' 32 through '~' 126)
func isPrintableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 32 || b > 126 { // excludes control chars and non-ASCII
			return false
		}
	}
	return true
}

// Register handles team registration by binding a team name to an existing code.
// Flow:
// - Expect POST with form fields: code, name
// - If code doesn't exist -> 404
// - If code already bound -> redirect to home (sets cookie)
// - Else create team, set cookie, redirect to home
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	form_code := r.FormValue("code")
	// Trim and normalize internal whitespace to a single space
	rawName := strings.TrimSpace(r.FormValue("name"))
	form_name := strings.Join(strings.Fields(rawName), " ")

	if form_code == "" {
		log.Printf("Register: missing code: code=%s, name=%s\n", form_code, form_name)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Código inválido")
		return
	}

	if form_name == "" {
		log.Printf("Register: missing name: code=%s, name=%s\n", form_code, form_name)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Nome do time deve conter apenas caracteres ASCII e ter no máximo 40 caracteres")
		return
	}

	// Validate team name: printable ASCII only and max length 40
	if len(form_name) > 40 || !isPrintableASCII(form_name) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Nome do time deve conter apenas caracteres ASCII e ter no máximo 40 caracteres")
		return
	}

	codes := controllers.NewCodeController(database.DB)
	teams := controllers.NewTeamController(database.DB)

	c, err := codes.FindCode(r.Context(), form_code)
	if err != nil {
		log.Println("Register: error getting code:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Código não encontrado")
		return
	}

	// Check if there is already a team bound to this code
	team, err := teams.GetTeamByCode(r.Context(), c.ID)
	if err != nil {
		log.Println("Register: error getting team by code:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	if team != nil {
		// Already bound -> behave like a successful login
		log.Println("Register: team already bound to code")
		setLoginCookieAndRedirect(w, r, form_code, "/")
		return
	}

	// Create the team
	_, err = teams.CreateTeam(r.Context(), form_name, c.ID)
	if err != nil {
		log.Println("Register: error creating the team:", err)

		var sqlite_err *sqlite.Error
		if errors.As(err, &sqlite_err) {
			if sqlite_err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, "Nome da equipe ou código já está em uso")
				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")

		return
	}

	// Set cookie and redirect to home
	setLoginCookieAndRedirect(w, r, form_code, "/")
}

// Helper to set the session cookie and redirect (supports HTMX via HX-Redirect)
func setLoginCookieAndRedirect(w http.ResponseWriter, r *http.Request, code, to string) {
	SetCookie(w, "team_code", code)
	SetHeader(w, "HX-Redirect", to)
	if GetHeader(r, "HX-Request") == "" {
		http.Redirect(w, r, to, http.StatusSeeOther)
	} else {
		fmt.Fprintf(w, "Redirecting to %s", to)
	}
}
