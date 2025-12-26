package handler

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/htetmyatthar/lothone/internal/config"
	"github.com/htetmyatthar/lothone/internal/utils"
	"github.com/htetmyatthar/lothone/middleware/csrf"
	"github.com/htetmyatthar/lothone/middleware/session"
	"github.com/htetmyatthar/lothone/web/components"
	"github.com/htetmyatthar/lothone/web/layout"
)

// Parse the dates into time.Time objects
const dateFormat = "2006-01-02"
const defaultAlterID = 1

func accountCreateHTMX(w http.ResponseWriter, r *http.Request) {
	log.Println("Account creation request received")

	username := r.FormValue("username")
	password := r.FormValue("password")
	accType := r.FormValue("type")
	serverId := r.FormValue("serverId")
	deviceId := r.FormValue("deviceId")
	sDate := r.FormValue("startDate")
	eDate := r.FormValue("endDate")

	log.Printf("Received form values - username: %s, password: %s, type: %s, serverId: %s, deviceId: %s, startDate: %s, endDate: %s", username, password, accType, serverId, deviceId, sDate, eDate)

	if deviceId == "" || sDate == "" || eDate == "" || username == "" || accType == "" {
		log.Println("Missing required fields in request")
		http.Error(w, "Invalid Request: missing required fields.", http.StatusBadRequest)
		return
	}

	if serverId == "" && password == "" {
		log.Println("Both serverId and password are missing")
		http.Error(w, "Invalid Request: both serverid and password can't be empty.", http.StatusBadRequest)
		return
	}

	if (serverId != "" && uuid.Validate(serverId) != nil) || (password != "" && uuid.Validate(password) != nil) {
		log.Println("Invalid UUID format in serverId or password")
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
		return
	}

	parsedAccType, err := utils.ParseAccountType(accType)
	if err != nil {
		log.Printf("Invalid account type: %s", accType)
		http.Error(w, "Invalid Request: invalid account type.", http.StatusBadRequest)
		return
	}

	if username == "-" {
		username = "unknown/admin"
		log.Println("Username set to default 'unknown/admin'")
	}

	startDate, err := time.Parse(dateFormat, sDate)
	if err != nil {
		log.Println("Invalid start date format")
		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse(dateFormat, eDate)
	if err != nil {
		log.Println("Invalid end date format")
		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
		return
	}

	if err := uuid.Validate(deviceId); err != nil {
		log.Println("Invalid device UUID")
		http.Error(w, "Invalid Request: invalid device uuid", http.StatusBadRequest)
		return
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("Failed to parse IP from RemoteAddr")
		http.Error(w, "Invalid request: unable to determine IP address", http.StatusBadRequest)
		return
	}

	newClient := utils.Client{
		Id:         serverId,
		AlterId:    defaultAlterID,
		Username:   username,
		DeviceId:   deviceId,
		StartDate:  strings.Split(startDate.String(), " ")[0],
		ExpireDate: strings.Split(endDate.String(), " ")[0],
		Password:   password,
	}

	log.Printf("Creating account of type: %s", parsedAccType)

	switch parsedAccType {
	case utils.SstpAccountType:
		d := r.FormValue("desc")
		if strings.Contains(username, "/") {
			log.Println("Invalid character '/' in SSTP username")
			http.Error(w, "Invalid username: please don't use '/' character inside sstp usernames.", http.StatusBadRequest)
			return
		}
		log.Println("Creating SSTP user...")
		resp, err := utils.CreateSSTPUser(newClient.Username, d, password, endDate)
		if err != nil {
			log.Printf("Failed to create SSTP user: %v", err)
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("SSTP creation response: %s", resp)

	case utils.ShadowsocksAccountType:
		log.Println("Creating Shadowsocks user...")
		c, u := utils.ShadowsocksAccountType.Filename()
		if err := utils.CreateShadowsocksUser(newClient, c, u); err != nil {
			log.Printf("Failed to create Shadowsocks user: %v", err)
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Restarting service after Shadowsocks user creation")
		if err := utils.RestartService(); err != nil {
			log.Printf("Failed to restart service: %v", err)
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case utils.VmessAccountType:
		log.Println("Creating Vmess user...")
		c, u := utils.VmessAccountType.Filename()
		if err := utils.CreateVmessUser(newClient, c, u); err != nil {
			log.Printf("Failed to create Vmess user: %v", err)
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Restarting service after Vmess user creation")
		if err := utils.RestartService(); err != nil {
			log.Printf("Failed to restart service: %v", err)
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Println("Sending Gotify notifications")
	title := *config.WebHost + " - New user is created"
	message := newClient.Username + "@" + *config.WebHostIP + " with [[" + newClient.Id + "]] is created by " + ip
	for _, key := range config.GotifyAPIKeys {
		utils.SendNoti(*config.GotifyServer, key, title, message, 5)
	}

	log.Println("Rendering success toast and refreshed account form")
	components.NotiToast("Account Created Successfully.").Render(context.Background(), w)
	components.AccountCreateForm(
		csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context())),
		templ.Attributes{"hx-swap-oob": "true"},
	).Render(context.Background(), w)
}

// accountCreateHTMX creates an account on the v2ray server and restart the v2ray service.
// func accountCreateHTMX(w http.ResponseWriter, r *http.Request) {
// 	var err error
// 	username, password, accType, serverId, deviceId, sDate, eDate := r.FormValue("username"), r.FormValue("password"), r.FormValue("type"), r.FormValue("serverId"), r.FormValue("deviceId"), r.FormValue("startDate"), r.FormValue("endDate")
// 	if deviceId == "" || sDate == "" || eDate == "" || username == "" || accType == "" {
// 		println(username, password, accType, serverId, deviceId, sDate, eDate)
// 		http.Error(w, "Invalid Request: missing required fields.", http.StatusBadRequest)
// 		return
// 	}
//
// 	// alert: if both of them's empty it's invalid.
// 	if serverId == "" && password == "" {
// 		http.Error(w, "Invalid Request: both serverid and password can't be empty.", http.StatusBadRequest)
// 		return
//
// 	}
//
// 	// alert: if either present one is invalid, it's an error
// 	if (serverId != "" && uuid.Validate(serverId) != nil) || (password != "" && uuid.Validate(password) != nil) {
// 		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
// 		return
// 	}
//
// 	parsedAccType, err := utils.ParseAccountType(accType)
// 	if err != nil {
// 		http.Error(w, "Invalid Request: invalid account type.", http.StatusBadRequest)
// 		return
// 	}
//
// 	// just a nice touch.
// 	if username == "-" {
// 		username = "unknown/admin"
// 	}
//
// 	var startDate time.Time
// 	var endDate time.Time
// 	startDate, err = time.Parse(dateFormat, sDate)
// 	if err != nil {
// 		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
// 		return
// 	}
//
// 	endDate, err = time.Parse(dateFormat, eDate)
// 	if err != nil {
// 		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
// 		return
// 	}
//
// 	err = uuid.Validate(deviceId)
// 	if err != nil {
// 		http.Error(w, "Invalid Request: invalid device uuid", http.StatusBadRequest)
// 		return
// 	}
//
// 	// doing things before writing to the file.
// 	var ip string
// 	ip, _, err = net.SplitHostPort(r.RemoteAddr)
// 	if err != nil {
// 		http.Error(w, "Invalid request: unable to determine IP address", http.StatusBadRequest)
// 		return
// 	}
//
// 	// modify the users by adding a modified user entity to the users file.
// 	newClient := utils.Client{
// 		Id:         serverId,
// 		AlterId:    defaultAlterID,
// 		Username:   username,
// 		DeviceId:   deviceId,
// 		StartDate:  strings.Split(startDate.String(), " ")[0],
// 		ExpireDate: strings.Split(endDate.String(), " ")[0],
// 		Password:   password,
// 	}
//
// 	switch parsedAccType {
// 	case utils.SstpAccountType:
// 		d := r.FormValue("desc")
// 		if strings.Contains(username, "/") {
// 			http.Error(w, "Invalid username: please don't use '/' character inside sstp usernames.", http.StatusBadRequest)
// 			return
// 		}
// 		resp, err := utils.CreateSSTPUser(newClient.Username, d, password, endDate)
// 		if err != nil {
// 			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		log.Println("response value from the sstp server: ", resp)
//
// 	case utils.ShadowsocksAccountType:
// 		c, u := utils.ShadowsocksAccountType.Filename()
// 		err = utils.CreateShadowsocksUser(newClient, c, u)
// 		if err != nil {
// 			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		err = utils.RestartService()
// 		if err != nil {
// 			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 	case utils.VmessAccountType:
// 		c, u := utils.VmessAccountType.Filename()
// 		err = utils.CreateVmessUser(newClient, c, u)
// 		if err != nil {
// 			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		err = utils.RestartService()
// 		if err != nil {
// 			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
//
// 	title := *config.WebHost + " - New user is created"
// 	message := newClient.Username + "@" + *config.WebHostIP + " with [[" + newClient.Id + "]] is created by " + ip
// 	for _, key := range config.GotifyAPIKeys {
// 		utils.SendNoti(*config.GotifyServer, key, title, message, 5)
// 	}
//
// 	components.NotiToast("Account Created Successfully.").Render(context.Background(), w)
// 	components.AccountCreateForm(
// 		csrf.Generate(
// 			w,
// 			"/accounts",
// 			session.GetSessionMgr().Token(r.Context())),
// 		templ.Attributes{"hx-swap-oob": "true"},
// 	).Render(context.Background(), w)
// 	return
// }

// accountQRGETHTMX gets qr data of specific account to show inside the modal.
func accountQRGETHTMX(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	t := r.FormValue("type")

	if idParam == "" || t == "" {
		http.Error(w, "Invalid Request: missing required fields.", http.StatusForbidden)
		return
	}

	err := uuid.Validate(idParam)
	if err != nil {
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusForbidden)
		return
	}

	accType, err := utils.ParseAccountType(t)
	if err != nil {
		http.Error(w, "Invalid Request: invalid account type.", http.StatusBadRequest)
		return
	}

	users, err := GetAllUsers(accType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user utils.Client
	if accType == utils.VmessAccountType {
		for _, user = range users {
			if user.Id == idParam {
				break
			}
		}
	} else {
		for _, user = range users {
			if user.Password == idParam {
				break
			}
		}
	}

	lk, lRemarks, err := GenerateLockedURI(user, accType)
	k, remarks, err := GenerateURI(user, accType)

	layout.QRTab(layout.QRData{
		Key:      lk,
		Username: user.Username + " " + user.DeviceId[len(user.DeviceId)-4:],
		Remarks:  lRemarks,
		Attributes: templ.Attributes{
			"id":          "lockedQRTab",
			"hx-swap-oob": "true",
		},
	}).Render(context.Background(), w)
	layout.QRTab(layout.QRData{
		Key:      k,
		Username: user.Username,
		Remarks:  remarks,
		Attributes: templ.Attributes{
			"id":          "openedQRTab",
			"hx-swap-oob": "true",
		},
	}).Render(context.Background(), w)
	return
}

// accountQRGETHTMX gets text data of specific account to show inside the modal.
func accountTextGETHTMX(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	t := r.FormValue("type")

	if idParam == "" || t == "" {
		http.Error(w, "Invalid Request: missing required fields.", http.StatusForbidden)
		return
	}

	err := uuid.Validate(idParam)
	if err != nil {
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusForbidden)
		return
	}

	accType, err := utils.ParseAccountType(t)
	if err != nil {
		http.Error(w, "Invalid Request: invalid account type.", http.StatusBadRequest)
		return
	}

	users, err := GetAllUsers(accType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user utils.Client
	if accType == utils.VmessAccountType {
		for _, user = range users {
			if user.Id == idParam {
				break
			}
		}
	} else {
		for _, user = range users {
			if user.Password == idParam {
				break
			}
		}
	}

	lk, _, err := GenerateLockedURI(user, accType)
	k, _, err := GenerateURI(user, accType)

	layout.TextKeyTab(layout.TextKeyData{
		Key: lk,
		Attributes: templ.Attributes{
			"id":          "lockedTextKeyTab",
			"hx-swap-oob": "true",
		},
	}).Render(context.Background(), w)
	layout.TextKeyTab(layout.TextKeyData{
		Key: k,
		Attributes: templ.Attributes{
			"id":          "openedTextKeyTab",
			"hx-swap-oob": "true",
		},
	}).Render(context.Background(), w)
	return
}

// accountDeleteHTMX deletes the account using the given server and device ids and restart the v2ray service.
func accountDeleteHTMX(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid request: unable to determine IP address", http.StatusBadRequest)
		return
	}
	serverId, deviceId, accType, username := r.FormValue("serverId"), r.FormValue("deviceId"), r.FormValue("type"), r.FormValue("username")

	if deviceId == "" || accType == "" {
		log.Println("this is the first 403")
		http.Error(w, "Invalid Request: missing required fields.", http.StatusForbidden)
		return
	}

	// alert: if both of them's empty it's invalid.
	if serverId == "" && username == "" {
		http.Error(w, "Invalid Request: both serverid and username can't be empty.", http.StatusBadRequest)
		return

	}

	// If serverId is provided but invalid, reject the request
	if serverId != "" && uuid.Validate(serverId) != nil {
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
		return
	}

	err = uuid.Validate(deviceId)
	if err != nil {
		log.Println("this is the device id error.")
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
		return
	}

	parsedAccType, err := utils.ParseAccountType(accType)
	if err != nil {
		log.Println("this is the acc type error.")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var status int
	var deletedUser *utils.Client
	switch parsedAccType {
	case utils.SstpAccountType:
		_, err = utils.DeleteSSTPUser(username)
		if err != nil {
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case utils.ShadowsocksAccountType:
		c, u := utils.ShadowsocksAccountType.Filename()
		deletedUser, status, err = utils.DeleteShadowsocksUser(serverId, deviceId, c, u)
		if err != nil {
			log.Println("Internal server error: ", err.Error())
			http.Error(w, "Internal Server Error: "+err.Error(), status)
			return
		}
		err = utils.RestartService()
		if err != nil {
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case utils.VmessAccountType:
		c, u := utils.VmessAccountType.Filename()
		deletedUser, status, err = utils.DeleteVmessUser(serverId, deviceId, c, u)
		if err != nil {
			log.Println("Internal server error: ", err.Error())
			http.Error(w, "Internal Server Error: "+err.Error(), status)
			return
		}
		err = utils.RestartService()
		if err != nil {
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Println("is this the error.")
	title := *config.WebHost + " - Existing user is deleted."
	var message string
	if parsedAccType != utils.SstpAccountType {
		message = deletedUser.Username + "@" + *config.WebHostIP + " with [[" + deletedUser.Id + "]] is deleted by " + ip
	} else {
		message = username + "@" + *config.WebHostIP + " SSTP server is deleted by " + ip
	}
	for _, key := range config.GotifyAPIKeys {
		utils.SendNoti(*config.GotifyServer, key, title, message, 5)
	}

	// NOTE: status 200 with empty response for successful deletion,
	// other status for failure to delete account.
	w.WriteHeader(http.StatusOK)
	utils.RestartService()
	return
}

func accountEditHTMX(w http.ResponseWriter, r *http.Request) {
	username, accType, deviceId, sDate, eDate := r.FormValue("username"), r.FormValue("type"), r.FormValue("deviceId"), r.FormValue("startDate"), r.FormValue("endDate")
	password, serverId := r.FormValue("password"), r.FormValue("serverId")

	if username == "" || accType == "" || deviceId == "" || sDate == "" || eDate == "" {
		http.Error(w, "Invalid Request: missing required fields.", http.StatusBadRequest)
		return
	}

	// alert: if both of them's empty it's invalid.
	if serverId == "" && password == "" {
		http.Error(w, "Invalid Request: both serverid and password can't be empty.", http.StatusBadRequest)
		return

	}

	parsedAccType, err := utils.ParseAccountType(accType)
	if err != nil {
		http.Error(w, "Invalid Request: invalid User Type", http.StatusBadRequest)
		return
	}

	// alert: if either present one is invalid, it's an error
	if (serverId != "" && uuid.Validate(serverId) != nil) || (password != "" && uuid.Validate(password) != nil) {
		http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
		return
	}

	if parsedAccType == utils.SstpAccountType {
		log.Println("Invoked unimplemented feature.")
		http.Error(w, "Account Edit Unavailable For SSTP Accounts", http.StatusNotImplemented)
	}

	// just a nice touch.
	if username == "-" {
		username = "unknown/admin"
	}

	startDate, err := time.Parse(dateFormat, sDate)
	if err != nil {
		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(dateFormat, eDate)
	if err != nil {
		http.Error(w, "Invalid Request: invalid date format", http.StatusBadRequest)
		return
	}

	err = uuid.Validate(deviceId)
	if err != nil {
		http.Error(w, "Invalid Request: invalid new device uuid", http.StatusBadRequest)
		return
	}

	// doing things before writing to the file.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid request: unable to determine IP address", http.StatusBadRequest)
		return
	}

	// modify the users by adding a modified user entity to the users file.
	modifiedClient := utils.Client{
		Id:         serverId,
		AlterId:    defaultAlterID,
		Username:   username,
		DeviceId:   deviceId,
		StartDate:  strings.Split(startDate.String(), " ")[0],
		ExpireDate: strings.Split(endDate.String(), " ")[0],
		Password:   password,
	}

	var oldClient *utils.Client
	var status int
	switch parsedAccType {
	case utils.VmessAccountType:
		cFile, uFile := parsedAccType.Filename()
		oldClient, status, err = utils.EditVmessUser(modifiedClient, cFile, uFile)
		if err != nil {
			http.Error(w, "Internal Server Error: "+err.Error(), status)
		}

		components.VmessAccount(
			modifiedClient,
			templ.Attributes{"hx-swap-oob": "true", "newly-swapped": "true"},
		).Render(context.Background(), w)

	case utils.ShadowsocksAccountType:
		cFile, uFile := parsedAccType.Filename()
		oldClient, status, err = utils.EditShadowsocksUser(modifiedClient, cFile, uFile)
		if err != nil {
			http.Error(w, "Internal Server Error: "+err.Error(), status)
		}

		components.ShadowsocksAccount(
			modifiedClient,
			templ.Attributes{"hx-swap-oob": "true", "newly-swapped": "true"},
		).Render(context.Background(), w)
	}

	title := *config.WebHost + " - User is updated"
	message := oldClient.Username + "@" + *config.WebHostIP + " with \nid: [[" + oldClient.Id + "]]\ndevice id: [[" + oldClient.DeviceId + "]]\n is updated by (" + ip + ") to " + modifiedClient.Username + "\ndevice id: [[" + modifiedClient.DeviceId + "]]"
	for _, key := range config.GotifyAPIKeys {
		utils.SendNoti(*config.GotifyServer, key, title, message, 5)
	}
	components.NotiToast("User information updated.").Render(context.Background(), w)
	return
}

func accountEditGetHTMX(w http.ResponseWriter, r *http.Request) {
	id, password := r.FormValue("serverId"), r.FormValue("password")
	t := r.FormValue("type")

	if t == "" {
		http.Error(w, "Invalid Request: missing required fields.", http.StatusBadRequest)
		return
	}

	// one of them must present.
	if id == "" && password == "" {
		http.Error(w, "Invalid Request: both serverid and password can't be empty.", http.StatusBadRequest)
		return
	}

	accType, err := utils.ParseAccountType(t)
	if err != nil {
		http.Error(w, "Invalid Request: invalid account type.", http.StatusBadRequest)
		return
	}

	// note: seperate this code block util the feature is implemented.
	if accType == utils.SstpAccountType {
		log.Println("SSTP is not yet implemented yet being called.")
		http.Error(w, "Not implemented", http.StatusNotImplemented)
		return
	}

	users, err := GetAllUsers(accType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// note: I just don't wanna compare two things at a time as
	// that can lead to perfomance overhead if there's more things to compare.
	found, user := false, utils.Client{}
	if accType == utils.VmessAccountType {
		log.Println("acc type is vmess")
		err := uuid.Validate(id)
		if err != nil {
			http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
			return
		}
		for _, c := range users {
			if c.Id == id {
				found, user = true, c
				break
			}
		}
	} else if accType == utils.ShadowsocksAccountType {
		log.Println("acc type is shadowsocks")
		err := uuid.Validate(password)
		if err != nil {
			http.Error(w, "Invalid Request: invalid UUID format.", http.StatusBadRequest)
			return
		}
		for _, c := range users {
			if c.Password == password {
				found, user = true, c
				break
			}
		}
	}

	if !found {
		log.Println("Invalid user is being searched.")
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(dateFormat, user.StartDate)
	if err != nil {
		http.Error(w, "Internal Server Error: invalid date format", http.StatusInternalServerError)
		return
	}

	expireDate, err := time.Parse(dateFormat, user.ExpireDate)
	if err != nil {
		http.Error(w, "Internal Server Error: invalid date format", http.StatusInternalServerError)
		return
	}

	// renders the user edit form.
	components.AccountEditForm(components.EditUserFormData{
		Username:   user.Username,
		DeviceUUID: user.DeviceId,
		ServerUUID: user.Id,
		StartDate:  startDate,
		EndDate:    expireDate,
		Password:   password,
		Type:       accType,
	},
		csrf.Generate(w, "/accounts", session.GetSessionMgr().Token(r.Context())),
	).Render(context.Background(), w) // BUG: gives out the csrf token.
	return
}
