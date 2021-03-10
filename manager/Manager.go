//Palette © Albert Bregonia 2021

package manager

import (
	"Palette/lobby"
	"log"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/sessions"
)

var (
	lobbies []*lobby.Data            //list of all lobbies
	timeOut = 30 * time.Second       //max time a lobby can be created yet still be empty
	add     = make(chan *lobby.Data) //channel for new lobby data
	mtx     = sync.RWMutex{}         //mutex to prevent data races
)

//LobbyManager deletes empty lobbies or adds new lobbies to the list of lobbies
func LobbyManager() {
	for {
		select {
		case new := <-add: //using the [add] channel ensures that [lobbies] isn't modified whilst being iterated through
			mtx.Lock()
			lobbies = append(lobbies, new)
			sort.Slice(lobbies, func(x, y int) bool { //quick sort
				return lobbies[x].Name() < lobbies[y].Name()
			})
			go new.Game().DataManager() //start the drawing data distribution thread
			go new.DataManager()        //start the chat data distribution thread
			mtx.Unlock()
		default: //garbage collection of lobby data
			for n := range lobbies {
				if t := time.Now().Sub(lobbies[n].TimeCreated()); len(lobbies[n].Players()) == 0 && t > timeOut { //if the lobby is old and empty
					lobbies[n].Shutdown() //shut down chat/drawing data threads
					log.Printf(`Lobby: '%v' has been deleted`, lobbies[n].Name())
					mtx.Lock()
					lobbies = append(lobbies[:n], lobbies[n+1:]...) //delete lobby from [lobbies]
					mtx.Unlock()
					runtime.GC() //force garbage collection
				}
			}
		}
	}
}

//LobbyHandler creates a new lobby based on given form data or creates cookie values that are valid for a lobby the user wishes to join
func LobbyHandler(w *http.ResponseWriter, r *http.Request, session *sessions.Session, store *sessions.CookieStore, newLobby bool) {
	if e := r.ParseForm(); e == nil {
		lobbyName, password, username := r.FormValue(`lobbyName`), r.FormValue(`lobbyPass`), r.FormValue(`username`)
		if existingLobby := GetLobby(lobbyName); existingLobby != nil { //checks if a lobby exists
			if existingLobby.Password() == password && !newLobby { //if the password is correct and trying to join
				existingLobby.UniqueName(&username) //ensure the username is unique
			} else {
				(*w).WriteHeader(http.StatusUnauthorized) //password incorrect
				return
			}
		} else if newLobby { //lobby doesn't exist and requesting to make a new lobby
			new := lobby.New(lobbyName, password, username)
			add <- &new //create new lobby and send data to LobbyManager() to pause the garbage collection
			log.Printf(`New Lobby Created: '%v'`, lobbyName)
		} else {
			(*w).WriteHeader(http.StatusNotFound) //trying to join a lobby that doesn't exist
			return
		}
		session.Values[`lobbyName`], session.Values[`lobbyPass`], session.Values[`username`] = lobbyName, password, username
		e = store.Save(r, *w, session) //save cookies
		if e != nil {
			http.Error(*w, e.Error(), http.StatusInternalServerError)
		} else {
			(*w).WriteHeader(http.StatusAccepted)
		}
	}
}

//GetLobby returns a pointer to a lobby given a name; [nil] if not found
func GetLobby(name string) *lobby.Data {
	mtx.RLock()
	defer mtx.RUnlock()
	for l, r := 0, len(lobbies)-1; l <= r; { //binary search
		mid := (l + r) / 2
		if lobbies[mid].Name() == name {
			return lobbies[mid]
		} else if lobbies[mid].Name() < name {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return nil
}
