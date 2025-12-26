package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/middleware/csrf"
	"github.com/htetmyatthar/lothone/middleware/session"
	"github.com/htetmyatthar/lothone/web/layout"
)

func dashboardSpecificHTMX(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")

	if t == "sstp" {
		log.Println("sstp dashboard is being rendered.")
		sstpUsers, err := utils.GetSSTPUsers()
		if err != nil {
			log.Println("sstp error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.SstpAccountsDashboard(sstpUsers, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		return
	}

	switch t {

	case "vmess":
		log.Println("vmess dashboard is being rendered.")
		users, err := GetAllUsers(utils.VmessAccountType)
		if err != nil {
			log.Println("vmess get users error: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.VmessAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		return

	case "shadowsocks":
		log.Println("shadowsocks dashboard is being rendered.")
		users, err := GetAllUsers(utils.ShadowsocksAccountType)
		if err != nil {
			log.Println("shadowsocks get users error: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.ShadowsocksAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		return
	}
}

func dashboardSpecificHTML(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")

	if t == "sstp" {
		log.Println("sstp dashboard is being rendered.")
		sstpUsers, err := utils.GetSSTPUsers()
		if err != nil {
			log.Println("sstp error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.AccountsDashboardPage(
			layout.SstpAccountsDashboard(sstpUsers, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))),
			csrf.Generate(w, "/logout", session.GetSessionMgr().Token(r.Context())),
		).Render(context.Background(), w)
		return
	}

	switch t {

	case "vmess":
		log.Println("vmess dashboard is being rendered.")
		users, err := GetAllUsers(utils.VmessAccountType)
		if err != nil {
			log.Println("vmess get users error: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.AccountsDashboardPage(
			layout.VmessAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))),
			csrf.Generate(w, "/logout", session.GetSessionMgr().Token(r.Context())),
		).Render(context.Background(), w)
		return

	case "shadowsocks":
		log.Println("shadowsocks dashboard is being rendered.")
		users, err := GetAllUsers(utils.ShadowsocksAccountType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.AccountsDashboardPage(
			layout.ShadowsocksAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))),
			csrf.Generate(w, "/logout", session.GetSessionMgr().Token(r.Context())),
		).Render(context.Background(), w)
		return
	}
}

func dashboardHTMX(w http.ResponseWriter, r *http.Request) {
	// get vmess user count
	vmessUsers, err := GetAllUsers(utils.VmessAccountType)
	if err != nil {
	}
	log.Println("VMESS USERS: ", len(vmessUsers))

	// get shadowsocks user count
	shadowsocksUsers, err := GetAllUsers(utils.ShadowsocksAccountType)
	if err != nil {
	}
	log.Println("Shadowsocks USERS: ", len(shadowsocksUsers))

	// get sstp user count
	sstpUsers, err := utils.GetSSTPUsers()
	if err != nil {
	}
	log.Println("sstp USERS: ", len(sstpUsers))

	layout.DashboardContent(
		csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context())),
		csrf.Generate(w, "/server/", session.GetSessionMgr().Token(r.Context())),
		strconv.Itoa(len(vmessUsers)+len(shadowsocksUsers)+len(sstpUsers)),
		strconv.Itoa(len(vmessUsers)),
		strconv.Itoa(len(shadowsocksUsers)),
		strconv.Itoa(len(sstpUsers)),
	).Render(context.Background(), w)
	return
}

func dashboardHTML(w http.ResponseWriter, r *http.Request) {
	// get vmess user count
	vmessUsers, err := GetAllUsers(utils.VmessAccountType)
	if err != nil {
	}
	log.Println("VMESS USERS: ", len(vmessUsers))

	// get shadowsocks user count
	shadowsocksUsers, err := GetAllUsers(utils.ShadowsocksAccountType)
	if err != nil {
	}
	log.Println("Shadowsocks USERS: ", len(shadowsocksUsers))

	// get sstp user count
	sstpUsers, err := utils.GetSSTPUsers()
	if err != nil {
	}
	log.Println("sstp USERS: ", len(sstpUsers))

	layout.DashboardPage(
		csrf.Generate(
			w,
			"/logout",
			session.GetSessionMgr().Token(r.Context())),
		csrf.Generate(
			w,
			"/accounts",
			session.GetSessionMgr().Token(r.Context())),
		strconv.Itoa(len(vmessUsers)+len(shadowsocksUsers)+len(sstpUsers)),
		strconv.Itoa(len(vmessUsers)),
		strconv.Itoa(len(shadowsocksUsers)),
		strconv.Itoa(len(sstpUsers)),
	).Render(context.Background(), w)
	return
}
