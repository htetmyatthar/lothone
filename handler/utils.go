// utils.go
//
// This file provides utility functions and structures for handling
// different types of HTTP requests in the `handler` package. It includes
// mechanisms for dynamically selecting and dispatching handlers based on
// request types such as HTML, HTMX, JSON, and XML.

// Key Components:
//
// Response Type Constants, muxHandler Struct, newMuxHandler, AddHandler Method,
// CreateHandler Method, defaultSelector Function, JSONRespond Function.
package handler

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/htetmyatthar/lothone/internal/config"
	"github.com/htetmyatthar/lothone/internal/utils"
)

// for differentiating the type of the request.
const (
	ResponseHTMX = iota
	ResponseHTML
	ResponseJSON
	ResponseXML
)

// to use in generating vpn URIs
const (
	VmessPrefix       = "vmess://"
	ShadowSocksPrefix = "shadowsocks://"
	V2boxLockedPrefix = "v2box://locked="
)

// muxHandler struct holds a list of handlers for different response type
type muxHandler struct {
	// handlers indexed by response type constants.
	handlers []http.HandlerFunc

	// selector checks and returns the different type of response handler for
	// different request types like, html, json, htmx, xml, etc.
	selector func(r *http.Request) int
}

// newMuxHandler creates a new
func newMuxHandler(selector func(r *http.Request) int, handlers ...http.HandlerFunc) *muxHandler {
	mh := &muxHandler{
		handlers: handlers,
		selector: selector,
	}
	return mh
}

// AddHandler appeds a handler to the `handlers` slice.
func (m *muxHandler) AddHandler(handler http.HandlerFunc) {
	m.handlers = append(m.handlers, handler)
}

// CreateHandler creates a different types of handlers for each type of requests
// using the `m.selector` function
func (m *muxHandler) CreateHandler() http.HandlerFunc {

	if m.selector == nil {
		log.Println("Using default selector for handling requests.")
		m.selector = defaultSelector
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// call the `selector` and returns the handler that is the most relevant
		index := m.selector(r)
		if index >= 0 && index < len(m.handlers) && m.handlers[index] != nil {
			m.handlers[index].ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid request type.", http.StatusNotImplemented)
		}

	}
}

// defaultSelector checks a common request types and returns the corresponding handler's index.
// default return response type is "application/json".
func defaultSelector(req *http.Request) int {
	// NOTE: since we are using the htmx for the website and we are targeting for it
	// htmx > html
	if req.Header.Get("HX-Request") == "true" {
		return ResponseHTMX
	}
	return ResponseHTML
}

// GenerateURI generates a usable URI for v2box application.
func GenerateURI(data utils.Client, t utils.AccountType) (key string, remarks string, err error) {
	key, remarks, err = "", "", errors.New("Not implemented")

	if t == utils.VmessAccountType {
		key, err = utils.GenerateVmessURI(data)
		remarks = strings.Split(*config.WebHost, ".")[0] + " " + data.Id[len(data.Id)-4:]
		return

	} else if t == utils.ShadowsocksAccountType {
		key, err = utils.GenerateShadowsocksURI(data)
		remarks = strings.Split(*config.WebHost, ".")[0] + " " + data.Password[len(data.Password)-4:]
		return
	}
	return
}

// GenerateURI generates a usable device id locked URI for v2box application.
func GenerateLockedURI(data utils.Client, t utils.AccountType) (key string, remarks string, err error) {
	key, remarks, err = "", "", errors.New("Not implemented")

	if t == utils.VmessAccountType {
		key, err = utils.GenerateVmessLockedURI(data)
		remarks = strings.Split(*config.WebHost, ".")[0] + " " + data.Id[len(data.Id)-4:]
		return

	} else if t == utils.ShadowsocksAccountType {
		key, err = utils.GenerateShadowsocksLockedURI(data)
		remarks = strings.Split(*config.WebHost, ".")[0] + " " + data.Password[len(data.Password)-4:]
		return
	}
	return
}

// GetAllUsers gets the users inside the users file.
// Suitable only for READ operations. Since you'll need other components for writing back.
// NOTE: This won't get sstp users.
func GetAllUsers(t utils.AccountType) ([]utils.Client, error) {
	if t == utils.SstpAccountType {
		return nil, errors.New("Use the GetSSTPUsers.")
	}

	_, filename := t.Filename()
	return loadAndGetUsers(filename)
}

// loadAndGetUsers is a function for getting users within a file.
// f must be a users file.
func loadAndGetUsers(f string) ([]utils.Client, error) {
	// load the users file.
	userData, err := os.ReadFile(f)
	if err != nil {
		log.Println("Error reading user data file: ", err)
		return nil, errors.New("Internal Server Error.")
	}

	// unmarshal the JSON users file into a map
	var userResult map[string]json.RawMessage
	err = json.Unmarshal(userData, &userResult)
	if err != nil {
		log.Println("Error unmarshalling JSON to map in users:", err)
		return nil, errors.New("Internal Server Error.")
	}

	// unmarshal the "users" key into a slice of clients.
	var users []utils.Client
	err = json.Unmarshal(userResult["clients"], &users)
	if err != nil {
		log.Println("Error unmarshalling 'users': ", err)
		return nil, errors.New("Internal Server Error.")
	}
	return users, nil
}
