package handler

import (
	"log"
	"net/http"

	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/middleware/session"
)

func logoutPOSTHTMX(w http.ResponseWriter, r *http.Request) {
	authenticated := session.GetSessionMgr().GetBool(r.Context(), utils.AuthenticatedField)

	if !authenticated {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// remove session and such?
	err := session.GetSessionMgr().Destroy(r.Context())
	if err != nil {
		log.Println("There's no session to be destroy.")
		return
	}

	w.Header().Set("HX-Redirect", "/login")
	w.Header().Set("HX-Pust-Url", "/login")
	w.WriteHeader(http.StatusOK)
	log.Println("Authenticated User Redirecting to login page and Loging Out.")
	return
}

func logoutPOSTHTML(w http.ResponseWriter, r *http.Request) {
	authenticated := session.GetSessionMgr().GetBool(r.Context(), utils.AuthenticatedField)

	if !authenticated {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// remove only the authentication field?
	session.GetSessionMgr().Put(r.Context(), utils.AuthenticatedField, false)

	http.Redirect(w, r, "/login", http.StatusFound)
	log.Println("Authenticated User Redirecting to login page and Loging Out.")
	return
}
