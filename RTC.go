// Palette © Albert Bregonia 2021
package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin: func(r *http.Request) bool {
			_, lobby, _ := ParseSession(nil, r)
			return lobby != nil //if the websocket connection is from a valid user in a lobby
		},
	}
	config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{`stun:stun.l.google.com:19302`}}},
	}
)

//SignalingSocket is a thread safe WebSocket used only for establishing WebRTC connections
type SignalingSocket struct {
	*websocket.Conn
	sync.RWMutex
}

//SendSignal is a thread safe wrapper for the `websocket.WriteJSON()` function that only sends the JSON form of a `Signal` struct
func (signaler *SignalingSocket) SendSignal(event, data string) error {
	signaler.Lock()
	defer signaler.Unlock()
	return signaler.WriteJSON(Signal{event, data})
}

//Signals to be written on a SignalingSocket in order to establish WebRTC connections
type Signal struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
