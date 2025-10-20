package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
	"github.com/secomp2025/localsnake/templates/pages"
)

var ADMIN_PASSWD = os.Getenv("ADMIN_PASSWD")

// LoginPage serves the login page component.
func LoginPage() http.Handler {
	return templ.Handler(pages.Login())
}

// LoginHandler dispatches by method on /login
// GET  -> render login page (redirects to / if already logged in)
// POST -> handle login submission (mocked routing)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetLogin(w, r)
	case http.MethodPost:
		PostLogin(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
	}
}

func GetLogin(w http.ResponseWriter, r *http.Request) {

	codes := controllers.NewCodeController(database.DB)
	teams := controllers.NewTeamController(database.DB)

	// If already logged in (valid cookie), go to home
	if c := GetCookieValue(r, "team_code"); c != "" {
		code, err := codes.FindCode(r.Context(), c)
		if err != nil {
			log.Println("LoginHandler: error getting code:", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
		if code == nil {
			ClearCookie(w, "team_code")
			LoginPage().ServeHTTP(w, r)
			return
		}

		team, err := teams.GetTeamByCode(r.Context(), code.ID)
		if err != nil {
			log.Println("LoginHandler: error getting team:", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
		if team == nil {
			templ.Handler(pages.Register(code.Code)).ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	LoginPage().ServeHTTP(w, r)
}

// PostLogin handles the HTMX form submission from the login page.
// Mocked behavior for now:
// - If the code is bound to a team (present in mockBound), go to the Home page
// - Otherwise, go to the registration page using the provided code
func PostLogin(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	form_code := r.FormValue("code")
	if form_code == "" {
		// Basic validation: re-render login with a minimal message (could be enhanced later)
		w.WriteHeader(http.StatusBadRequest)
		templ.Handler(pages.Login()).ServeHTTP(w, r)
		return
	}

	form_code = strings.ToUpper(form_code)

	log.Println("LoginHandler: code =", form_code)

	codes := controllers.NewCodeController(database.DB)
	teams := controllers.NewTeamController(database.DB)

	code, err := codes.FindCode(r.Context(), form_code)
	if err != nil {
		log.Println("LoginHandler: error getting code:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	if code == nil {
		log.Println("LoginHandler: code not found")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Código não encontrado")
		return
	}

	team, err := teams.GetTeamByCode(r.Context(), code.ID)
	if err != nil {
		log.Println("LoginHandler: error getting team:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	if team == nil {
		log.Println("LoginHandler: team not found")
		// send to registration page
		templ.Handler(pages.Register(form_code)).ServeHTTP(w, r)
		return
	}

	if team.IsAdmin.Valid && team.IsAdmin.Bool {
		// require password
		templ.Handler(pages.PasswordLogin(team.ID)).ServeHTTP(w, r)
		return
	}

	// Bound -> set cookie and redirect to home
	http.SetCookie(w, &http.Cookie{
		Name:     "team_code",
		Value:    form_code,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// For HTMX, prefer HX-Redirect header
	w.Header().Set("HX-Redirect", "/")
	// Also set a standard redirect status for non-htmx fallbacks
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		// Minimal body for HTMX (not used if HX-Redirect is honored)
		fmt.Fprintf(w, "Redirecting to home for %s", team.Name)
	}
}

func PostLoginAdm(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	team_id := r.FormValue("team_id")
	if team_id == "" {
		// Basic validation: re-render login with a minimal message (could be enhanced later)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Team ID not provided")
		return
	}
	team_id_int, err := strconv.ParseInt(team_id, 10, 64)
	if err != nil {
		log.Println("LoginHandler: error parsing team ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	teams := controllers.NewTeamController(database.DB)
	team, err := teams.GetTeamByCode(r.Context(), team_id_int)
	if err != nil {
		log.Println("LoginHandler: error getting team:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	if team == nil {
		log.Println("LoginHandler: team not found")
		// send to registration page
		templ.Handler(pages.Login()).ServeHTTP(w, r)
		return
	}

	form_password := r.FormValue("password")
	if form_password != ADMIN_PASSWD {
		log.Println("LoginHandler: password not correct")
		// Basic validation: re-render login with a minimal message (could be enhanced later)
		w.WriteHeader(http.StatusBadRequest)
		templ.Handler(pages.PasswordLogin(team_id_int)).ServeHTTP(w, r)
		return
	}

	// Bound -> set cookie and redirect to home
	http.SetCookie(w, &http.Cookie{
		Name:     "team_code",
		Value:    "ADMBSNAKE",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// For HTMX, prefer HX-Redirect header
	w.Header().Set("HX-Redirect", "/")
	// Also set a standard redirect status for non-htmx fallbacks
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		// Minimal body for HTMX (not used if HX-Redirect is honored)
		fmt.Fprintf(w, "Redirecting to home for %s", team.Name)
	}
}
