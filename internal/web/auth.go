package web

import (
	"crypto/subtle"
	"net/http"
)

const settingsCookie = "ikite_settings"

func (s *Server) authorizeSettings(w http.ResponseWriter, r *http.Request) bool {
	pass := s.Cfg.SettingsPass
	if pass == "" {
		http.NotFound(w, r)
		return false
	}

	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("pass")), []byte(pass)) == 1 {
		http.SetCookie(w, &http.Cookie{
			Name:     settingsCookie,
			Value:    pass,
			Path:     "/settings",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   365 * 24 * 3600,
			Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		})
		return true
	}

	if c, err := r.Cookie(settingsCookie); err == nil {
		if subtle.ConstantTimeCompare([]byte(c.Value), []byte(pass)) == 1 {
			return true
		}
	}

	http.NotFound(w, r)
	return false
}
