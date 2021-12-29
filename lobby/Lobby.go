// Palette © Albert Bregonia 2021
package lobby

import (
	"Palette/lobby/user"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

/*
	Lobby is a container that manages a collection of users.

	Lobby utilizes a thread safe map that uses usernames as keys and uses pointers to `User` instances as the values.
	Similar to `Manager`, upon creation, the constructor will start up the `userManager()` goroutine. This goroutine
	will poll the users in the map and see if the user has exceeded the lobby's maximum timeout period and delete their
	data from the lobby.
*/
type Lobby struct {
	name, password string
	users          map[string]*user.User
	host           *user.User
	chat           chan Message
	shutdown       chan string //channel to signal the manager to delete, should only be accessed by manager
	maxTimeout     time.Duration
	sync.RWMutex
}

// === Lobby Properties === //

//Constructor for a lobby object, starts the newly created lobby's `userManager()` goroutine.
//NOTE: `host` cannot be `nil`, this function will panic if so as it will initialize the map of users with `{host.Name(): host}`
func New(name, password string, maxTimeout time.Duration, host *user.User) *Lobby {
	lobby := Lobby{
		name:       name,
		password:   password,
		users:      map[string]*user.User{host.Name(): host},
		host:       host,
		chat:       make(chan Message),
		maxTimeout: maxTimeout,
		RWMutex:    sync.RWMutex{},
	}
	go lobby.userManager()
	return &lobby
}

// Accessors (a pointer is used to prevent copying the struct)

//Name is an accessor for a lobby's name value
func (lobby *Lobby) Name() string {
	lobby.RLock()
	defer lobby.RUnlock()
	return lobby.name
}

//Password is an accessor for a lobby's password value
func (lobby *Lobby) Password() string {
	lobby.RLock()
	defer lobby.RUnlock()
	return lobby.password
}

//Host is an accessor for a lobby's host user.
//The host has elevated permissions and is the only one allowed to reconfigure any settings
func (lobby *Lobby) Host() *user.User {
	lobby.RLock()
	defer lobby.RUnlock()
	return lobby.host
}

//Size is an accessor for a lobby's number of users (active and inactive)
func (lobby *Lobby) Size() int {
	lobby.RLock()
	defer lobby.RUnlock()
	return len(lobby.users)
}

//Chat is an accessor for for a lobby's chat message channel. It is immutable
func (lobby *Lobby) Chat() chan Message { return lobby.chat }

// Mutators

//SetName is a mutator for a lobby's name value
func (lobby *Lobby) SetName(name string) {
	lobby.Lock()
	defer lobby.Unlock()
	lobby.name = name
}

//SetPassword is a mutator for a lobby's password value
func (lobby *Lobby) SetPassword(password string) {
	lobby.Lock()
	defer lobby.Unlock()
	lobby.password = password
}

//SetHost is a mutator for the host of a lobby given the new host's username.
//Returns an error if a user with the given username has not joined the lobby
func (lobby *Lobby) SetHost(name string) error {
	lobby.Lock()
	defer lobby.Unlock()
	if lobby.users[name] == nil {
		return fmt.Errorf(
			`unable to change host of '%v': '%v' has not joined this lobby`,
			lobby.name, name,
		)
	}
	lobby.host = lobby.users[name]
	return nil
}

// === User management === //

//GetUser is an accessor for a pointer to a specific user in a lobby given their username. External use only!
func (lobby *Lobby) GetUser(name string) *user.User {
	lobby.RLock()
	defer lobby.RUnlock()
	return lobby.users[name]
}

//AddUser adds a pointer to a user to the `users` map of a lobby.
//If the given user has a name that is not unqiue relative to the lobby, it will be adjusted.
//Returns an error if the pointer given is `nil`. External use only!
func (lobby *Lobby) AddUser(user *user.User) error {
	lobby.Lock()
	defer lobby.Unlock()
	return lobby.addUser(user)
}

//addUser is the mutex free version of AddUser(). Internal use only!
func (lobby *Lobby) addUser(user *user.User) error {
	if user == nil {
		return fmt.Errorf(`cannot add 'nil' as a user to lobby: '%v'`, lobby.name)
	}
	name := user.Name()
	newName := name
	//i+1 bc adjusted names will start at 'name-1' instead of 'name-0'
	for i := 0; lobby.users[newName] != nil; i++ {
		newName = fmt.Sprintf(`%v-%v`, name, i+1)
	}
	if newName != name { //don't wait for mutex release if we don't have to
		user.SetName(newName)
		name = newName
	}
	lobby.users[name] = user
	// log.Printf(`[%v] '%v' has joined.`, lobby.name, name)
	return nil
}

//RemoveUser removes a pointer to user data to the `users` map of a lobby.
//Returns an error if a user with the given name is not found in the lobby. External Use only!
func (lobby *Lobby) RemoveUser(name string) error {
	lobby.Lock()
	defer lobby.Unlock()
	return lobby.removeUser(name)
}

//removeUser is the mutex free version of RemoveUser(). Internal use only!
func (lobby *Lobby) removeUser(name string) error {
	user := lobby.users[name]
	if user == nil {
		return fmt.Errorf(
			`unable to delete user: '%v' from lobby: '%v'. '%v' has not joined this lobby`,
			name, lobby.name, name,
		)
	}
	delete(lobby.users, name)
	// log.Printf(`[%v] Player data for '%v' was deleted.`, lobby.name, name)
	if user == lobby.host {
		for _, u := range lobby.users {
			lobby.host = u //pick a random person in the map to be the next host
			// log.Printf(`[%v] New host: '%v'`, lobby.name, u.Name())
			break
		}
		if lobby.host == user { //if the host didn't change after the loop, set it to nil, required for garbage collection
			lobby.host = nil
		}
	}
	return nil
}

//Message represents a chat message to be broadcasted to every user or a specific user in a lobby
type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Time    string `json:"time"`
}

//userManager is a goroutine that handles distributing data to users and user data deletion.
//If the lobby is empty, this goroutine will shutdown and signal the manager to delete it
func (lobby *Lobby) userManager() {
	for {
		select {
		case msg, open := <-lobby.chat:
			if !open {
				return
			}
			bin, _ := json.Marshal(msg)
			lobby.Lock()
			for _, user := range lobby.users {
				chat := user.Channel(`chat`)
				if chat != nil { //skip user if they are trying to reconnect
					chat.SendText(string(bin))
				}
			}
			lobby.Unlock()
		default: //delete old users after lobby.maxTimeout
			lobby.Lock()
			for _, User := range lobby.users {
				if User.TimeDisconnect() != user.NIL_TIME && time.Since(User.TimeDisconnect()) >= lobby.maxTimeout {
					lobby.removeUser(User.Name())
				}
			}
			if len(lobby.users) == 0 {
				close(lobby.chat)            //shutdown this goroutine
				lobby.shutdown <- lobby.name //signal the manager to delete this lobby
			}
			lobby.Unlock()
		}
	}
}
