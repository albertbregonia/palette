// Palette © Albert Bregonia 2021
package user

import (
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

//Unix epoch to be used as a `nil` value for time
var NIL_TIME time.Time = time.Unix(0, 0)

/*
	Manages a single user's data and connection to the server.

	As a user is required to have a single WebSocket connection for WebRTC signaling and multiple WebRTC DataChannels in order
	to interact with the users in a lobby, a `User` is a named representation of those connections.
*/
type User struct {
	name       string
	disconnect time.Time
	channels   map[string]*webrtc.DataChannel //map of WebRTC data channels based on their label
	attributes map[string]interface{}
	sync.RWMutex
}

// === User Properties === //

//Constructor for a User
func New(name string) *User {
	return &User{
		name:       name,
		disconnect: NIL_TIME,
		channels:   make(map[string]*webrtc.DataChannel),
		attributes: make(map[string]interface{}),
		RWMutex:    sync.RWMutex{},
	}
}

// Accessors (a pointer is used to prevent copying the struct)

//Name is an accessor for a user's name value
func (user *User) Name() string {
	user.RLock()
	defer user.RUnlock()
	return user.name
}

//TimeDisconnect is an accessor for a user's time of disconnect
func (user *User) TimeDisconnect() time.Time {
	user.RLock()
	defer user.RUnlock()
	return user.disconnect
}

//Channel is an accessor for a channel in a user's map of WebRTC data channels given a label
func (user *User) Channel(label string) *webrtc.DataChannel {
	user.RLock()
	defer user.RUnlock()
	return user.channels[label]
}

// Mutators

//SetName is a mutator for a user's name value
func (user *User) SetName(name string) error {
	user.Lock()
	defer user.Unlock()
	if len(name) == 0 || name == `Palette` {
		return fmt.Errorf(`invalid name, username length must be > 0 and cannot be 'Palette'`)
	}
	user.name = name
	return nil
}

//SetTimeDisconnect is a mutator for a user's time of disconnect
func (user *User) SetTimeDisconnect(time time.Time) {
	user.Lock()
	defer user.Unlock()
	user.disconnect = time
}

//SetChannel is a mutator for a channel in a user's map of WebRTC data channels given a label and channel pointer
func (user *User) SetChannel(label string, channel *webrtc.DataChannel) {
	user.Lock()
	defer user.Unlock()
	user.channels[label] = channel
}
