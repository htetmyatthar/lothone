package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/middleware/csrf"
	"github.com/htetmyatthar/lothone/middleware/session"
	"github.com/htetmyatthar/lothone/web/components"
	"github.com/htetmyatthar/lothone/web/layout"
)

func dashboardSpecificRefreshHTMX(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")

	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
		return
	}

	switch t {

	case "sstp":
		log.Println("sstp dashboard is being rendered.")
		sstpUsers, err := utils.GetSSTPUsers()
		if err != nil {
			log.Println("sstp error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.SstpAccountsDashboard(sstpUsers, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		components.NotiToast("SSTP dashboard refreshed.").Render(context.Background(), w)

	case "vmess":
		log.Println("vmess dashboard is being rendered.")
		users, err := GetAllUsers(utils.VmessAccountType)
		if err != nil {
			log.Println("vmess get users error: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.VmessAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		components.NotiToast("Vmess dashboard refreshed.").Render(context.Background(), w)

	case "shadowsocks":
		log.Println("shadowsocks dashboard is being rendered.")
		users, err := GetAllUsers(utils.ShadowsocksAccountType)
		if err != nil {
			log.Println("shadowsocks get users error: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		layout.ShadowsocksAccountsDashboard(users, csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context()))).Render(context.Background(), w)
		components.NotiToast("Shadowsocks dashboard refreshed.").Render(context.Background(), w)

	default:
		w.Header().Set("HX-Redirect", "/dashboard")
		w.Header().Set("HX-Push-Url", "/dashboard")
		w.WriteHeader(http.StatusFound)
	}
}
