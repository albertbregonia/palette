//Palette © Albert Bregonia 2021

package player

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//Data describes all the data a player object holds
type Data struct {
	name          string
	points        int
	timeDC        time.Time
	connection    *Connection
	*sync.RWMutex //*Required to read/write lock and prevent data races
}

//New is the default constructor for a player
func New(name string, connection *websocket.Conn) Data {
	mtx := sync.RWMutex{}
	newConnection := Connection{connection, nil, &mtx}
	return Data{name, 0, time.Now(), &newConnection, &mtx}
}

//Name is an accessor for the name of the given player instance
func (player *Data) Name() string {
	player.RLock()
	defer player.RUnlock()
	return player.name
}

//SetName is a mutator for the name value of the given player instance
func (player *Data) SetName(name string) *Data {
	player.Lock()
	player.name = name
	player.Unlock()
	return player
}

//Points is an accessor for the points value of the given player instance
func (player *Data) Points() int {
	player.RLock()
	defer player.RUnlock()
	return player.points
}

//SetPoints is a mutator for the points value of the given player instance
func (player *Data) SetPoints(points int) *Data {
	player.Lock()
	player.points = points
	player.Unlock()
	return player
}

//TimeDC is an accessor for the time of disconnect for the given player instance
func (player *Data) TimeDC() time.Time {
	player.RLock()
	defer player.RUnlock()
	return player.timeDC
}

//SetTimeDC is a mutator for the time of disconnect for the given player instance
func (player *Data) SetTimeDC(time time.Time) *Data {
	player.Lock()
	player.timeDC = time
	player.Unlock()
	return player
}

//Connection is an accessor for the connection value of the given player instance
func (player *Data) Connection() *Connection {
	player.RLock()
	defer player.RUnlock()
	return player.connection
}

//SetConnection is a mutator for the connection value of the given player instance
func (player *Data) SetConnection(connection *Connection) *Data {
	player.Lock()
	player.connection = connection
	player.Unlock()
	return player
}
