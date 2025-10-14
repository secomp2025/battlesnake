package handlers

import "net/http"

func SetCookie(w http.ResponseWriter, name string, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetCookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func SetHeader(w http.ResponseWriter, name string, value string) {
	w.Header().Set(name, value)
}

func GetHeader(r *http.Request, name string) string {
	return r.Header.Get(name)
}
