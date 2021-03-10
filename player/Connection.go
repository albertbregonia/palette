//Palette © Albert Bregonia 2021

package player

import (
	"sync"

	"github.com/gorilla/websocket"
)

//Connection is an object representing the two sockets for chat data and drawing data
type Connection struct {
	chat, draw *websocket.Conn
	*sync.RWMutex
}

//Chat is an accessor for the chat websocket
func (c *Connection) Chat() *websocket.Conn {
	c.RLock()
	defer c.RUnlock()
	return c.chat
}

//SetChat is a mutator for the chat websocket
func (c *Connection) SetChat(chat *websocket.Conn) *Connection {
	c.Lock()
	defer c.Unlock()
	c.chat = chat
	return c
}

//Draw is an accessor for the draw server websocket
func (c *Connection) Draw() *websocket.Conn {
	c.RLock()
	defer c.RUnlock()
	return c.draw
}

//SetDraw is a mutator for the draw server websocket
func (c *Connection) SetDraw(draw *websocket.Conn) *Connection {
	c.Lock()
	defer c.Unlock()
	c.draw = draw
	return c
}
