// Palette © Albert Bregonia 2021
package lobby

import (
	"log"
	"sync"
)

/*
	Manager is a container that manages a collection of lobbies.

	Manager utilizes a thread safe map that uses lobby names as keys and uses pointers to `Lobby` instances as the values.
	Upon creation, the constructor will start up the `cleanup()` goroutine that will handle signals to delete a lobby given
	its name; similar to an interrupt.
*/
type Manager struct {
	lobbies  map[string]*Lobby
	shutdown chan string
	sync.RWMutex
}

//Constructor for a lobby manager object that starts the newly created lobby manager's cleanup() goroutine
func NewManager() *Manager {
	manager := Manager{
		lobbies:  make(map[string]*Lobby),
		shutdown: make(chan string),
		RWMutex:  sync.RWMutex{},
	}
	go manager.cleanup()
	return &manager
}

//cleanup is to be used as a separate goroutine. It handles a manager's `shutdown` channel
//and will block until it is signaled with the name of a lobby to delete.
func (manager *Manager) cleanup() {
	for {
		lobbyName, open := <-manager.shutdown
		if !open {
			return
		}
		manager.Lock()
		delete(manager.lobbies, lobbyName)
		manager.Unlock()
		log.Printf(`[Manager] Lobby: '%v' has been deleted`, lobbyName)
	}
}

//AddLobby adds a pointer to a lobby to a manager's `lobbies` map
func (manager *Manager) AddLobby(lobby *Lobby) {
	manager.Lock()
	defer manager.Unlock()
	manager.lobbies[lobby.Name()] = lobby
	lobby.shutdown = manager.shutdown
	log.Printf(`[Manager] Lobby: '%v' has been created by '%v'`, lobby.Name(), lobby.Host().Name())
}

//GetLobby is an accessor for a lobby in a manager's `lobbies` map given a name
func (manager *Manager) GetLobby(name string) *Lobby {
	manager.RLock()
	defer manager.RUnlock()
	return manager.lobbies[name]
}
