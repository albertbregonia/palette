//Palette © Albert Bregonia 2021

package game

import (
	"Palette/player"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//Data describes all the data a game object holds
type Data struct {
	live                      bool
	duration, rounds, current int
	host, word                string
	artistData                *player.Data
	words                     []string
	artists                   []*player.Connection
	paint                     chan *interface{}
	add                       chan *player.Connection
	remove                    chan int
	quit                      chan bool
	sync.RWMutex
}

//New is the default constructor for a game
func New(host string) Data {
	return Data{
		false, 80, 3, 0, host, ``, nil,
		make([]string, 0),
		make([]*player.Connection, 0),
		make(chan *interface{}),
		make(chan *player.Connection),
		make(chan int),
		make(chan bool),
		sync.RWMutex{}}
}

//Live is an accessor that returns a boolean on whether or not given game instance is currently running
func (game *Data) Live() bool {
	game.RLock()
	defer game.RUnlock()
	return game.live
}

//SetLive is a mutator that sets a boolean on whether or not given game instance is currently running
func (game *Data) SetLive(live bool) *Data {
	game.Lock()
	game.live = live
	game.Unlock()
	return game
}

//Shutdown shuts down the DataManger thread for the on-going game
func (game *Data) Shutdown() {
	log.Printf(`Draw Data Manager thread for [%v]'s lobby has been shutdown.`, game.host)
	game.quit <- true
}

//Duration is an accessor for the round duration value of the given game instance
func (game *Data) Duration() int {
	game.RLock()
	defer game.RUnlock()
	return game.duration
}

//SetDuration is a mutator for the round duration value of the given game instance
func (game *Data) SetDuration(duration int) *Data {
	game.duration = duration
	return game
}

//Rounds is an accessor for the number of rounds of the given game instance
func (game *Data) Rounds() int {
	game.RLock()
	defer game.RUnlock()
	return game.rounds
}

//SetRounds is a mutator for the number of rounds of the given game instance
func (game *Data) SetRounds(rounds int) *Data {
	game.rounds = rounds
	return game
}

//Host is an accessor for the host name value of the given game instance
func (game *Data) Host() string {
	game.RLock()
	defer game.RUnlock()
	return game.host
}

//SetHost is a mutator for the host name value of the given game instance
func (game *Data) SetHost(host string) *Data {
	game.RLock()
	defer game.RUnlock()
	game.host = host
	return game
}

//Word is an accessor for the current word value of the given game instance
func (game *Data) Word() string {
	game.RLock()
	defer game.RUnlock()
	return game.word
}

//SetWord is a mutator for the current word value of the given game instance
func (game *Data) SetWord(word string) *Data {
	game.word = word
	return game
}

// ChooseWord is a mutator for the current word value of the given game instance.
// This function randomly selects from the word list the new value of [word]
func (game *Data) ChooseWord() *Data {
	game.Lock()
	rand.Seed(time.Now().UnixNano())
	game.word = game.words[rand.Intn(len(game.words))]
	game.Unlock()
	return game
}

//Words is an accessor for the  words array value of the given game instance
func (game *Data) Words() []string {
	game.RLock()
	defer game.RUnlock()
	return game.words
}

//SetWords is a mutator for the [words] array value of the given game instance
func (game *Data) SetWords(words *[]string) *Data {
	game.words = *words
	return game
}

// AddWord is a mutator for the given game instance's [words] array. Given a pointer to another string, this function
// will add that word to the end of the [words] array.
func (game *Data) AddWord(word *string) *Data {
	game.words = append(game.words, *word)
	return game
}

// AddWords is a mutator for the given game instance's [words] array. Given a pointer to another string array, this function
// will iterate through that list and add each word to the [words] array.
func (game *Data) AddWords(words *[]string) *Data {
	for _, word := range *words {
		game.words = append(game.words, word)
	}
	return game
}

// RemoveWord is a mutator for the given game instance's [words] array. Given a pointer to another string, this function
// will iterate through the [words] array until it finds the desired string and removes it.
func (game *Data) RemoveWord(word *string) *Data {
	for n := range game.words {
		if strings.EqualFold(game.words[n], *word) {
			game.words = append(game.words[:n], game.words[n+1:]...)
			break
		}
	}
	return game
}

// RemoveWords is a mutator for the given game instance's [words] array. Given a pointer to another string array, this function
// will iterate through that list and remove each word from the [words] array if found.
func (game *Data) RemoveWords(words *[]string) *Data {
	for _, word := range *words {
		for n := range game.words {
			if strings.EqualFold(game.words[n], word) {
				game.words = append(game.words[:n], game.words[n+1:]...)
			}
		}
	}
	return game
}

//Paint is an accessor for the channel to send JSON drawing data to the DataManager thread
func (game *Data) Paint() chan *interface{} {
	game.RLock()
	defer game.RUnlock()
	return game.paint
}

//Artists returns the artist position given a connection pointer; Returns [-1] if not found
func (game *Data) Artists() []*player.Connection {
	game.RLock()
	defer game.RUnlock()
	return game.artists
}

//AddArtist adds a connection to the list of players to send drawing data to;
func (game *Data) AddArtist(connection *player.Connection) {
	game.add <- connection
}

//RemoveArtist removes a connection to the list of players to send drawing data to
func (game *Data) RemoveArtist(n int) *Data {
	game.remove <- n
	return game
}

//CurrentArtist returns the current artist to draw; returns [nil] if the artist list is empty
func (game *Data) CurrentArtist() *player.Connection {
	game.RLock()
	defer game.RUnlock()
	if len(game.artists) > 0 {
		return game.artists[game.current]
	}
	return nil
}

//ArtistData returns the current artist player data; returns nil if not set
func (game *Data) ArtistData() *player.Data {
	game.RLock()
	defer game.RUnlock()
	return game.artistData
}

//SetArtistData returns the current artist player data; returns nil if not set
func (game *Data) SetArtistData(player *player.Data) *player.Data {
	game.Lock()
	game.artistData = player
	game.Unlock()
	return player
}

//SetCurrentArtist sets the current artist to draw; returns [nil] if the artist list is empty or [n] is out of range
func (game *Data) SetCurrentArtist(n int) *player.Connection {
	game.Lock()
	defer game.Unlock()
	if n > 0 && n < len(game.artists) {
		game.current = n
		return game.artists[game.current]
	}
	return nil
}

//NextArtist returns the next current artist to draw; returns [nil] if the artist list is empty
func (game *Data) NextArtist() *player.Connection {
	game.Lock()
	defer game.Unlock()
	game.current++
	if len(game.artists) > 0 {
		if game.current >= len(game.artists) { //circular
			game.current = 0
			return game.artists[0]
		}
		return game.artists[game.current]
	}
	game.current = 0
	return nil
}

//GetArtist returns the artist position given a connection pointer; Returns [-1] if not found
func (game *Data) GetArtist(connection *websocket.Conn) int {
	game.RLock()
	defer game.RUnlock()
	for n := range game.artists {
		if game.artists[n].Draw() == connection {
			return n
		}
	}
	return -1
}

//Hint randomly changes underscores to the corresponding letter in the current value of [word]
func (game *Data) Hint(word string) string {
	rand.Seed(time.Now().UnixNano())
	hint := ``
	pos := rand.Intn(len(word))
	for i, char := range word {
		if i == pos {
			if char == '_' {
				hint += string(game.word[i])
			} else {
				pos = rand.Intn(len(word) - i)
				hint += string(char)
			}
		} else {
			hint += string(char)
		}
	}
	return hint
}

//NHints calls the Hint() function [n] amount of times
func (game *Data) NHints(word string, n int) string {
	hint := word
	for i := n; i < n; i++ {
		hint = game.Hint(hint)
	}
	return hint
}

// DataManager is the main thread to redistribute received drawing data to players in a lobby
// similar to a lobby object's DataManager, DataManager also handles the adding/removal of players
func (game *Data) DataManager() {
	for {
		select {
		case <-game.quit:
			return
		case n := <-game.remove:
			if n >= 0 && n < len(game.artists) {
				game.Lock()
				game.artists = append(game.artists[:n], game.artists[n+1:]...)
				game.Unlock()
				if n == game.current {
					game.NextArtist()
				}
			}
		case artist := <-game.add:
			game.Lock()
			game.artists = append(game.artists, artist)
			game.Unlock()
		default: //lowest priority
			select {
			case data := <-game.paint:
				for n := 0; n < len(game.artists); n++ {
					game.RLock()
					if n != game.current {
						_ = game.artists[n].Draw().WriteJSON(*data)
					}
					game.RUnlock()
				}
			default:
			}
		}
	}
}
