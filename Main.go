// Palette © Albert Bregonia 2021
package main

import (
	"Palette/lobby"
	"Palette/lobby/user"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// The Main package handles the web server backend and WebRTC connections

var (
	MAX_USER_TIMEOUT = 5 * time.Minute
	manager          = lobby.NewManager()
)

var (
	//go:embed frontend/*
	embedded  embed.FS
	secureKey = securecookie.GenerateRandomKey(512)
	key       = fmt.Sprintf("%d%d%d%d", //generate random keys per server
		rand.Intn(math.MaxInt), rand.Intn(math.MaxInt),
		rand.Intn(math.MaxInt), rand.Intn(math.MaxInt))
	store = sessions.NewCookieStore(secureKey) //valid cookie sessions
)

func main() {
	frontend, _ := fs.Sub(embedded, `frontend`)
	http.Handle(`/`, http.FileServer(http.FS(frontend)))
	http.HandleFunc(`/reconnect`, ReconnectHandler)
	http.HandleFunc(`/login`, LoginHandler)
	http.HandleFunc(`/leave`, LeaveLobby)
	log.Println(`Palette Web Server Initialized`)
	log.Fatal(http.ListenAndServeTLS(`:443`, `server.crt`, `server.key`, nil))
}

//ParseSession parses the cookies of a request and returns the user's session, lobby and username
//If any values are invalid, (nil, nil, ``) is returned
func ParseSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, *lobby.Lobby, string) {
	session, e := store.Get(r, key)
	if e != nil {
		if w != nil {
			http.Error(w, e.Error(), http.StatusNotFound)
		}
		return nil, nil, ``
	}
	lobby := manager.GetLobby(fmt.Sprint(session.Values[`lobby`]))
	if lobby == nil {
		if w != nil {
			http.Error(w, `lobby not found`, http.StatusNotFound)
		}
		return nil, nil, ``
	}
	return session, lobby, fmt.Sprint(session.Values[`username`])
}

//ReconnectHandler checks if a user has already joined a lobby and is merely reconnecting
func ReconnectHandler(w http.ResponseWriter, r *http.Request) {
	_, lobby, username := ParseSession(w, r)
	if lobby == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	usr := lobby.GetUser(username)
	usr.SetTimeDisconnect(user.NIL_TIME) //disable data deletion timer for this user
	fmt.Fprintf(w, `You have been reconnected to '%v' as '%v'`, lobby.Name(), username)
}

//LoginHandler handles a request to create/join a lobby and establishes the required cookies for the user
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	//validate request values
	lobbyName := strings.TrimSpace(r.FormValue(`lobby`))
	password := strings.TrimSpace(r.FormValue(`password`))
	username := strings.TrimSpace(r.FormValue(`username`))
	createLobby, _ := strconv.ParseBool(r.FormValue(`create`))
	for _, parameter := range []string{lobbyName, password, username} {
		if parameter == `` {
			http.Error(w, `one or more required parameters were empty`, http.StatusBadRequest)
			return
		}
	}
	if username == `Palette` {
		http.Error(w, `Invalid Username. This name is reserved.`, http.StatusConflict)
		return
	}
	//perform request operation
	existingLobby := manager.GetLobby(lobbyName)
	if createLobby { //making a lobby
		if existingLobby != nil { //lobby already exists
			w.WriteHeader(http.StatusConflict)
			return
		}
		manager.AddLobby(lobby.New(lobbyName, password, MAX_USER_TIMEOUT, user.New(username)))
	} else { //joining a lobby
		if existingLobby == nil { //lobby does not exist
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if existingLobby.Password() != password { //password is wrong
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user := user.New(username)
		existingLobby.AddUser(user)
		username = user.Name()
	}
	//save valid session to cookies
	session, _ := store.Get(r, key)
	session.Values[`lobby`] = lobbyName
	session.Values[`username`] = username
	store.Save(r, w, session)
	w.WriteHeader(http.StatusAccepted)
}

//LeaveLobby handles a deliberate request for a user to leave a lobby and have their data deleted
func LeaveLobby(w http.ResponseWriter, r *http.Request) {
	session, lobby, username := ParseSession(w, r)
	if session == nil {
		return //nothing to do, they already don't have data
	}
	if e := lobby.RemoveUser(username); e != nil { //delete user data
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1 //delete cookie from user
	store.Save(r, w, session)   //delete cookie from valid cookies
}
