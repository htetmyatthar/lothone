package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/web/components"
)

func accountFormGet(w http.ResponseWriter, r *http.Request) {
	accountType := r.FormValue("type")
	if accountType == "" {
		log.Println("account type is empty")
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	t, err := strconv.Atoi(accountType)
	if err != nil {
		log.Println("account type is not being parsed correctly", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	switch utils.AccountType(t) {
	case utils.VmessAccountType:
		components.VmessAccountCreate().Render(context.Background(), w)
	case utils.ShadowsocksAccountType:
		components.ShadowsocksAccountCreate().Render(context.Background(), w)
	case utils.SstpAccountType:
		components.SstpAccountCreate().Render(context.Background(), w)
	}
	return
}
