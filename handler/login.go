package handler

import (
	"context"
	"log"
	"net"
	"net/http"
	// "os"
	"strings"

	// "github.com/goccy/go-json"
	"github.com/htetmyatthar/lothone/internal/config"
	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/middleware/csrf"
	"github.com/htetmyatthar/lothone/middleware/session"
	"github.com/htetmyatthar/lothone/web/layout"
)

func loginHTMX(w http.ResponseWriter, r *http.Request) {
	// HACK: use the exists method instead of getbool?
	authenticated := session.GetSessionMgr().GetBool(r.Context(), utils.AuthenticatedField)

	if authenticated {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		log.Println("Authenticated User Redirecting to dashboard")
		return
	}

	session.GetSessionMgr().Put(r.Context(), utils.AuthenticatedField, false)
	token, _, _ := session.GetSessionMgr().Commit(r.Context()) // note: Commit() method also checks that a session exists or not.
	t := csrf.Generate(w, "/login", token)

	layout.LoginMain(t, config.Version).Render(context.Background(), w)
	return
}

func loginHTML(w http.ResponseWriter, r *http.Request) {
	authenticated := session.GetSessionMgr().GetBool(r.Context(), utils.AuthenticatedField)

	if authenticated {
		log.Println("Authenticated User Redirecting to dashboard")
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	session.GetSessionMgr().Put(r.Context(), utils.AuthenticatedField, false)
	token, _, _ := session.GetSessionMgr().Commit(r.Context()) // note: Commit() method also checks that a session exists or not.
	t := csrf.Generate(w, "/login", token)

	layout.LoginPage(t, config.Version).Render(context.Background(), w)
	return
}

func loginPOSTHTMX(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("IP not found to log error.")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// new token for error form.
	t := csrf.Generate(w, "/login", session.GetSessionMgr().Token(r.Context()))

	name := r.FormValue("username")
	pw := r.FormValue("password")

	if name == "" || pw == "" {
		log.Println("Attempt with blank credentials")
		layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
		return
	}

	// NOTE: sanity check. internal/utils.go:216 func InitPanelUsers
	// BUG: refactor this so that the insanity check will pass through and changing one value won't affect other
	// initpaneluser and this should use the same function to check the validation of username and password.
	if len(name) > 30 || len(pw) > 30 {
		layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
		return
	}

	if _, ok := utils.PanelUsers[name]; !ok {
		log.Println("Attempt with wrong username.")
		layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
		return
	}

	correct, err := utils.VerifyPassword(pw, utils.PanelUsers[name])
	// handle hashing errors.
	if err != nil && err != utils.ErrWrongPassword {
		log.Println("verifying user password gone wrong.", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !correct {
		log.Println("Attempt with wrong password.", pw)
		// send a notification to the gotify server.
		title := *config.WebHost + " - " + name + " logged in"
		message := name + " logged into " + *config.WebHostIP + " using wrong password and " + ip
		for _, key := range config.GotifyAPIKeys {
			utils.SendNoti(*config.GotifyServer, key, title, message, 9)
		}
		layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
		return
	}

	// NOTE: below this assumes user is authenticated.
	rememberMe := r.FormValue("remember")
	if rememberMe == "1" { // checked.
		session.GetSessionMgr().RememberMe(r.Context(), true)
	}

	session.GetSessionMgr().Put(r.Context(), utils.AuthenticatedField, true)

	// send a notification to the gotify server.
	title := *config.WebHost + " - " + name + " logged in"
	message := name + " logged into " + *config.WebHostIP + " and " + ip
	for _, key := range config.GotifyAPIKeys {
		utils.SendNoti(*config.GotifyServer, key, title, message, 9)
	}

	url := session.GetSessionMgr().GetString(r.Context(), utils.URLAfterLogin)
	if strings.Contains(url, "dashboard") {
		w.Header().Set("HX-Push-Url", url)
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusFound)
		session.GetSessionMgr().Remove(r.Context(), utils.URLAfterLogin)
		return
	}
	w.Header().Set("HX-Push-Url", "/dashboard")
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusFound)
	session.GetSessionMgr().Remove(r.Context(), utils.URLAfterLogin)
	return
}

// HACK: not finished yet.
// func loginPOSTHTML(w http.ResponseWriter, r *http.Request) {
// 	ip, _, err := net.SplitHostPort(r.RemoteAddr)
// 	if err != nil {
// 		log.Println("IP not found to log error.")
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// t := csrf.Generate(w, "/login", m.GetSessionMgr().Token(r.Context()), "")
//
// 	name := r.FormValue("username")
// 	pw := r.FormValue("password")
//
// 	if name == "" || pw == "" {
// 		log.Println("Attempt with blank credentials")
// 		// BUG: redirect back to the login form?
// 		// layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
// 		return
// 	}
//
// 	// NOTE: sanity check. internal/utils.go:196
// 	if len(name) > 20 || len(pw) > 30 {
// 		// BUG: redirect back to the login form?
// 		// layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
// 		return
// 	}
//
// 	if _, ok := utils.PanelUsers[name]; !ok {
// 		log.Println("Attempt with wrong username.")
// 		// BUG: redirect back to the login form?
// 		// layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
// 		return
// 	}
//
// 	correct, err := utils.VerifyPassword(pw, utils.PanelUsers[name])
// 	// handle hashing errors.
// 	if err != nil && err != utils.ErrWrongPassword {
// 		log.Println("verifying user password gone wrong.", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	if !correct {
// 		log.Println("Attempt with wrong password.", pw)
// 		// send a notification to the gotify server.
// 		title := *config.WebHost + " - " + name + " logged in"
// 		message := name + " logged into " + *config.WebHostIP + " using wrong password and " + ip
// 		for _, key := range config.GotifyAPIKeys {
// 			utils.SendNoti(*config.GotifyServer, key, title, message, 9)
// 		}
// 		// BUG: redirect back to the login form?
// 		// layout.LoginFormWithError(t, name, pw).Render(context.Background(), w)
// 		return
// 	}
//
// 	// NOTE: below this assumes user is authenticated.
// 	rememberMe := r.FormValue("remember")
// 	if rememberMe == "1" { // checked.
// 		session.GetSessionMgr().RememberMe(r.Context(), true)
// 	}
//
// 	session.GetSessionMgr().Put(r.Context(), utils.AuthenticatedField, true)
//
// 	// load the users file.
// 	userData, err := os.ReadFile(*config.UserFile)
// 	if err != nil {
// 		log.Println("Error reading user data file: ", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// unmarshal the JSON users file into a map
// 	var userResult map[string]json.RawMessage
// 	err = json.Unmarshal(userData, &userResult)
// 	if err != nil {
// 		log.Println("Error unmarshalling JSON to map in users:", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// unmarshal the "users" key into a slice of clients.
// 	var users []utils.Client
// 	err = json.Unmarshal(userResult["clients"], &users)
// 	if err != nil {
// 		log.Println("Error unmarshalling 'users': ", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// send a notification to the gotify server.
// 	title := *config.WebHost + " - " + name + " logged in"
// 	message := name + " logged into " + *config.WebHostIP + " and " + ip
// 	for _, key := range config.GotifyAPIKeys {
// 		utils.SendNoti(*config.GotifyServer, key, title, message, 9)
// 	}
//
// 	layout.DashboardPage(
// 		csrf.Generate(
// 			w,
// 			"/logout",
// 			session.GetSessionMgr().Token(r.Context())),
// 		csrf.Generate(
// 			w,
// 			"/accounts",
// 			session.GetSessionMgr().Token(r.Context())),
// 	).Render(context.Background(), w)
// 	return
// }
