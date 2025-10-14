package handlers

import (
	"fmt"
	"net/http"
)

// Logout clears the team_code cookie and redirects to /login
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "team_code",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
