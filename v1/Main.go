//Palette © Albert Bregonia 2021

package main

import (
	"Palette/manager"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	secureKey = securecookie.GenerateRandomKey(512)
	key       = fmt.Sprintf("%16d%16d%16d%16d", //generate random keys per session
		rand.Intn(10000000000000000),
		rand.Intn(10000000000000000),
		rand.Intn(10000000000000000),
		rand.Intn(10000000000000000))
	store      = sessions.NewCookieStore(secureKey) //cookie jar
	wsUpgrader = websocket.Upgrader{                //websocket upgrader
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func main() {
	wsUpgrader.CheckOrigin = verifyOrigin
	http.HandleFunc(`/game`, hasGame)
	http.HandleFunc(`/new`, newLobby)
	http.HandleFunc(`/join`, joinLobby)
	http.HandleFunc(`/leave`, leaveLobby)
	http.HandleFunc(`/chat`, chatHandler)
	http.HandleFunc(`/draw`, drawHandler)

	go manager.LobbyManager() //garbage collector for empty lobbies, handles lobby creation as well
	log.Println(`[Palette] Server has started`)
	http.Handle(`/`, http.FileServer(http.Dir(`frontend`)))
	log.Fatal(http.ListenAndServe(`:54000`, nil)) //start server on port 54000
}

// hasGame() checks if the player has the cookies of a valid lobby.
// Returns 202 if the cookies are of a valid lobby, 401 otherwise
func hasGame(w http.ResponseWriter, r *http.Request) {
	session, e := store.Get(r, key)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	if lobby := manager.GetLobby(fmt.Sprint(session.Values[`lobbyName`])); lobby != nil {
		w.WriteHeader(http.StatusAccepted) //202
	} else {
		w.WriteHeader(http.StatusUnauthorized) //401
	}
}

//newLobby() forwards form data to the lobby manager to make a new lobby if possible
func newLobby(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, key)
	manager.LobbyHandler(w, r, session, store, true)
}

//joinLobby() forwards form data to the lobby manager to join a lobby if possible
func joinLobby(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, key)
	manager.LobbyHandler(w, r, session, store, false)
}

// leaveLobby() only deletes the player's cookies.
// The player data will be handled by the lobby's DataManager() thread
func leaveLobby(w http.ResponseWriter, r *http.Request) {
	session, e := store.Get(r, key)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	e = store.Save(r, w, session)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

//chatHandler() calls wsHandler() to take valid cookies and establish a websocket connection for chat
func chatHandler(w http.ResponseWriter, r *http.Request) {
	wsHandler(w, r, true)
}

//drawHandler() calls wsHandler() to take valid cookies and establish a websocket connection for drawing data
func drawHandler(w http.ResponseWriter, r *http.Request) {
	wsHandler(w, r, false)
}

//verifyOrigin ensures that a websocket connection request is from a valid player
func verifyOrigin(r *http.Request) bool {
	session, _ := store.Get(r, key)
	lobby := manager.GetLobby(fmt.Sprint(session.Values[`lobbyName`]))
	return lobby != nil
}

// wsHandler() checks for valid cookies and establishes a websocket connection
// and a respective data reader thread for either chat data or drawing data
func wsHandler(w http.ResponseWriter, r *http.Request, chat bool) {
	session, _ := store.Get(r, key)
	username, lobbyName := fmt.Sprint(session.Values[`username`]), fmt.Sprint(session.Values[`lobbyName`])
	if lobby := manager.GetLobby(lobbyName); lobby != nil {
		if connection, e := wsUpgrader.Upgrade(w, r, nil); e == nil {
			if chat {
				lobby.ChatReader(connection, username)
			} else {
				lobby.PaintLoader(connection, username)
			}
		} else {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	}
}
