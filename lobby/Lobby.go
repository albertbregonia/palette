//Palette © Albert Bregonia 2021

package lobby

import (
	"Palette/game"
	"Palette/player"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//Data describes all the data a lobby object holds
type Data struct {
	name, password   string
	game             *game.Data
	players          []*player.Data
	add              chan *player.Data
	chat             chan string
	quit, stop, next chan bool
	creation         time.Time
	sync.RWMutex     //*Required to read/write lock and prevent data races
}

const (
	serverPrefix     = `<b style='color: red'>Server:</b> `
	messageAll       = `<anmt>`
	currentArtistTag = `<draw>`
	colorPositive    = `var(--highlight2)`
	colorNegative    = `var(--highlight3)`
	maxTimeOut       = 5 * time.Minute
)

//New is the default constructor for a lobby
func New(name, password, host string) Data {
	game := game.New(host)
	return Data{name, password, &game,
		make([]*player.Data, 0),
		make(chan *player.Data),
		make(chan string),
		make(chan bool),
		make(chan bool),
		make(chan bool),
		time.Now(),
		sync.RWMutex{}}
}

//Name is an accessor for the name value of the given lobby instance
func (lobby *Data) Name() string {
	return lobby.name
}

//Password is an accessor for the password value of the given lobby instance
func (lobby *Data) Password() string {
	return lobby.password
}

//Players is an accessor for the list of players in the given lobby instance
func (lobby *Data) Players() []*player.Data {
	lobby.Lock()
	defer lobby.Unlock()
	return lobby.players
}

//GetPlayer returns a player index and a pointer to the player data given a name; [-1, nil] if the player is not found
func (lobby *Data) GetPlayer(name string) (int, *player.Data) {
	lobby.RLock()
	defer lobby.RUnlock()
	for l, r := 0, len(lobby.players)-1; l <= r; { //binary search
		mid := (l + r) / 2
		if lobby.players[mid].Name() == name {
			return mid, lobby.players[mid]
		} else if lobby.players[mid].Name() < name {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return -1, nil
}

//Game is an accessor for the game data of the given lobby instance
func (lobby *Data) Game() *game.Data {
	return lobby.game
}

//TimeCreated is an accessor for the time that the given lobby instance was created
func (lobby *Data) TimeCreated() time.Time {
	return lobby.creation
}

// Disconnect is the event handler for when a player disconnects from the lobby.
// Given a player, it notifies lobby, records the time of disconnect and deletes their connection data
// to queue the deletion of the whole player object. Once [maxTimeOut] has elapsed and the player still
// has not reconnected, the player object is deleted by the DataManager thread
// [notify] is a boolean that represents whether or not to notify the lobby of the disconnect
func (lobby *Data) Disconnect(player *player.Data, notify bool, deliberate bool) {
	if notify {
		lobby.messageAll(colorNegative, player.Name()+` has left the lobby`)
	}
	if deliberate {
		player.SetTimeDC(time.Now().Add(-(maxTimeOut + time.Second)))
	} else {
		player.SetTimeDC(time.Now())
	}
	player.Connection().SetChat(nil)
}

//Shutdown shuts down the Data Manager thread for the given lobby instance
func (lobby *Data) Shutdown() {
	log.Printf(`Main Data Manager thread for [%v] has been shutdown.`, lobby.name)
	lobby.quit <- true
	if lobby.game.Live() {
		log.Printf(`Game Data Manager thread for [%v] has been shutdown.`, lobby.name)
		lobby.stop <- true
	}
}

// DataManager is the main thread that handles all data for the given lobby instance.
// It controls new players being added to the lobby, old players being removed from the lobby,
// and the echoing of chat messages to all players in the lobby.
// This thread runs infinitely until lobby.Shutdown() is called.
func (lobby *Data) DataManager() {
	for {
		select {
		case <-lobby.quit:
			return
		default:
			select {
			case new := <-lobby.add: //pause chat echo and player cleanup to add new players to list
				lobby.Lock()
				lobby.players = append(lobby.players, new)
				sort.Slice(lobby.players, func(x, y int) bool { //sort
					return lobby.players[x].Name() < lobby.players[y].Name()
				})
				lobby.Unlock()
			case msg := <-lobby.chat: //send chat messages to players in the given lobby instance
				data := []byte(msg)
				if strings.Index(msg, currentArtistTag) > -1 { //message current artist only
					_ = lobby.game.CurrentArtist().Chat().WriteMessage(websocket.TextMessage, data)
				} else { //broadcast message to all users
					for n := 0; n < len(lobby.players); n++ {
						if connection := lobby.players[n].Connection().Chat(); connection != nil {
							_ = connection.WriteMessage(websocket.TextMessage, data) //no handling upon error as it would be handled by the reader thread
						}
					}
				}
			default: //player data garbage collection
				for n := 0; n < len(lobby.players); n++ {
					player := lobby.players[n] //players have 5 minutes to reconnect after disconnecting from the lobby before their data is deleted
					if time.Now().Sub(player.TimeDC()) > maxTimeOut && player.Connection().Chat() == nil {
						log.Printf(`Deleted player data for: %v`, player.Name())
						lobby.Lock()
						lobby.players = append(lobby.players[:n], lobby.players[n+1:]...)
						lobby.Unlock()
						if player.Name() == lobby.game.Host() && len(lobby.players) >= 1 { //change host if possible
							log.Printf(`Changed Host for [%v] to [%v]`, lobby.name, lobby.players[0].Name())
							lobby.game.SetHost(lobby.players[0].Name())
						}
						runtime.GC()
					}
				}
			}
		}
	}
}

//ChatReader is the main thread in which messages are received and commands are parsed from the front end
func (lobby *Data) ChatReader(connection *websocket.Conn, username string) {
	_, p := lobby.GetPlayer(username)
	if p == nil { //make new player if they don't exist
		newPlayer := player.New(username, connection) //player data
		p = &newPlayer
		lobby.add <- p
	}
	for p == nil { //wait until the player data is added to the lobby
		_, p = lobby.GetPlayer(username)
	}
	p.Connection().SetChat(connection)
	lobby.messageAll(colorPositive, username+` has joined the lobby.`)
	for { //handle messages until disconnect
		if _, rawData, e := connection.ReadMessage(); e == nil { //send message down broadcast channel
			msg := string(rawData)
			switch msg[:1] {
			case `;`: // ; is the command prefix
				lobby.ParseCommand(username == lobby.game.Host(), msg[1:], connection) //only the host can perform certain commands
			case fmt.Sprintf(`%c`, 8): //deliberate disconnect
				lobby.Disconnect(p, true, true)
				return //[return] is used here to end the thread as [break] will only leave the switch
			default: //send standard messages with a time stamp, highlighted name and message
				if strings.EqualFold(msg, lobby.game.Word()) && lobby.game.Live() {
					if connection != lobby.game.CurrentArtist().Chat() {
						lobby.messageAll(`limegreen`, username+` has guessed the word correctly!`)
						p.SetPoints(p.Points() + 100)
					}
				} else {
					lobby.chat <- fmt.Sprintf(`<b title='%v' style='color: %v'>%v</b> %v`, game.Now(), colorPositive, username, msg)
				}
			}
		} else { //disconnect and end thread upon error
			lobby.Disconnect(p, true, false)
			break
		}
	}
}

//PaintLoader is the main thread in which drawing data is received as JSON and sent to the GameManager to be redistributed
func (lobby *Data) PaintLoader(connection *websocket.Conn, username string) {
	_, p := lobby.GetPlayer(username) //error checking is not included as the front end JavaScript ensures that the player exists
	p.Connection().SetDraw(connection)
	lobby.game.AddArtist(p.Connection())
	for {
		var data interface{}
		if e := connection.ReadJSON(&data); e == nil { //send incoming drawing data down the pipeline otherwise disconnect
			if lobby.game.CurrentArtist().Draw() == connection {
				lobby.game.Paint() <- &data
			}
		} else if e.Error() != `invalid character 'o' looking for beginning of value` { //only disconnects upon fatal error
			lobby.game.RemoveArtist(lobby.game.GetArtist(connection))
			lobby.Disconnect(p, false, false)
			if p.Connection().Chat() != nil { //if the chat is connected but an error occurs here; notify the lobby
				lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight3)'>%v has encountered an error. Please ask them to refresh.</b>`, serverPrefix, p.Name())
			}
			break
		}
	}
}

//Pictionary *this should be moved to the game package*
func (lobby *Data) runGame() {
	select {
	case <-lobby.stop:
		lobby.game.SetLive(false)
		return
	default:
		for i := 3; i > 0; i-- {
			lobby.chat <- fmt.Sprintf(`%v<b style='color: '>Game Starting in...%v</b>`, messageAll, i)
			time.Sleep(time.Second)
		}
		for round := 0; round < lobby.game.Rounds(); round++ {
			lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>Round: %v</b>`, serverPrefix, round+1)
			for n := 0; n < len(lobby.game.Artists()); n++ {
				max, _ := time.ParseDuration(fmt.Sprintf(`%vs`, lobby.game.Duration()))
				lobby.game.ChooseWord()
				lobby.chat <- fmt.Sprintf(`%v%v<b style='color: limegreen'>Your word is: <b>%v</b></b>`, currentArtistTag, serverPrefix, lobby.game.Word())
				hint := game.Hide(lobby.game.Word())
				lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>Hint: %v (%v)</b>`, messageAll, game.Space(hint), len(lobby.game.Word()))
				for timeLeft := max; timeLeft > 0; timeLeft -= time.Second {
					select {
					case <-lobby.next:
						timeLeft = 0
						continue
					default:
						if timeLeft == max || timeLeft%time.Minute == 0 || timeLeft == 30*time.Second || timeLeft <= 10*time.Second {
							lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>%v remaining.</b>`, serverPrefix, timeLeft)
						}
						if timeLeft != max && timeLeft%(30*time.Second) == 0 {
							newHint := lobby.game.NHints(lobby.game.Hint(hint), 3)
							for newHint == hint { //ensure decent hints
								newHint = lobby.game.NHints(lobby.game.Hint(hint), 3)
							}
							hint = newHint
							lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>New Hint: %v (%v)</b>`, messageAll, game.Space(hint), len(lobby.game.Word()))
						}
						time.Sleep(time.Second)
					}
				}
				lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>Round Over. The word was <b style='color: limegreen'>%v</b></b>`, messageAll, lobby.game.Word())
				time.Sleep(5 * time.Second)
				lobby.game.NextArtist()
			}
		}
		lobby.chat <- fmt.Sprintf(`%v<b style='color: var(--highlight2)'>Game Over. Please type ;players to see your scores!`, messageAll)
		lobby.game.SetCurrentArtist(0)
		lobby.game.SetLive(false)
	}
}

//ParseCommand is the main controller for handling commands sent by a lobby
func (lobby *Data) ParseCommand(host bool, cmd string, connection *websocket.Conn) {
	args := getArgs(cmd)
	msg := ``
	if len(args) > 0 {
		switch args[0] {
		case `time`:
			if len(args) == 2 && host {
				if amt, e := strconv.Atoi(args[1]); e == nil && amt >= 30 && amt <= 180 {
					lobby.Game().SetDuration(amt)
					msg = fmt.Sprintf(`Round duration: <b style='color: limegreen'> %v seconds</b>`, amt)
				}
			}
		case `rounds`:
			if len(args) == 2 && host {
				if amt, e := strconv.Atoi(args[1]); e == nil && amt >= 1 && amt <= 10 {
					lobby.Game().SetRounds(amt)
					msg = fmt.Sprintf(`Number of rounds: <b style='color: limegreen'>%v</b>`, amt)
				}
			}
		case `players`:
			players := `List of Players:<br>`
			lobby.RLock()
			for _, p := range lobby.players {
				players += fmt.Sprintf(`<b>%v<b> %v<br>`, p.Points(), p.Name())
			}
			lobby.RUnlock()
			msg += players
		case `words`:
			if len(args) > 1 && host { //the host in only allowed to make modifications
				switch len(args) {
				case 2:
					if args[1] == `clear` {
						empty := make([]string, 0)
						lobby.game.SetWords(&empty)
						msg = `Successfully cleared word list.`
					}
				case 3:
					switch args[1] {
					case `set`:
						words, wordList := parseWords(&args[2])
						lobby.game.SetWords(words)
						msg = fmt.Sprintf(`Successfully updated word list: <b style='color: limegreen'>%v</b>`, *wordList)
					case `add`:
						lobby.game.AddWord(&args[2])
						msg = fmt.Sprintf(`Successfully added word to word list: <b style='color: limegreen'>%v</b>`, args[2])
					case `add-all`:
						words, wordList := parseWords(&args[2])
						lobby.game.AddWords(words)
						msg = fmt.Sprintf(`Successfully added words to word list: <b style='color: limegreen'>%v</b>`, *wordList)
					case `remove`:
						lobby.game.RemoveWord(&args[2])
						msg = fmt.Sprintf(`Successfully removed word from word list: <b style='color: limegreen'>%v</b>`, args[2])
					case `remove-all`:
						words, wordList := parseWords(&args[2])
						lobby.game.RemoveWords(words)
						msg = fmt.Sprintf(`Successfully removed words from word list: <b style='color: limegreen'>%v</b>`, *wordList)
					}
				}
			} else { //non-hosts can only list the words
				wordList := ``
				for i, word := range lobby.game.Words() {
					wordList += fmt.Sprintf(`%v. "%v"<br>`, i, word)
				}
				msg = fmt.Sprintf(`Word list: <br><b style='color: var(--highlight2)'>%v</b>`, wordList)
			}
		case `random-words`:
			msg = `Successfully generated random words for word list [TO BE IMPLEMENTED]`
		case `start`:
			if host && len(lobby.game.Words()) > 10 && !lobby.game.Live() {
				lobby.game.SetLive(true)
				go lobby.runGame()
			} else {
				msg = `<b style='color: red'>Unable to start game. Please check your settings. Is the game already running?</b>`
			}
		case `next`:
			if lobby.game.Live() && len(lobby.next) == 0 {
				if host || connection == lobby.game.CurrentArtist().Chat() {
					msg = `Changing to next player...`
					lobby.next <- true
				}
			}
		case `draw`:
			if host && !lobby.game.Live() {
				lobby.message(currentArtistTag, colorPositive, `<b style='color: limegreen'>Your word is: <b>Draw.</b> You are now able to freely draw.</b>`)
			}
		}
	}
	if msg != `` { //if a valid message is queued
		lobby.messageAll(colorPositive, msg)
	}
}

//UniqueName loops and increments iteration count until the given name is no longer a duplicate
func (lobby *Data) UniqueName(name *string) {
	lobby.RLock()
	newName := *name
	i := 1
	for _, p := range lobby.players {
		for p.Name() == newName {
			newName = fmt.Sprintf(`%v-%v`, *name, i)
			i++
		}
	}
	*name = newName
	lobby.RUnlock()
}

func (lobby *Data) messageAll(color, msg string) { //messageAll to server
	lobby.chat <- fmt.Sprintf(`%v<b title='%v' style='color: %v'>%v</b>`, serverPrefix, game.Now(), color, msg)
}

func (lobby *Data) message(tag, color, msg string) { //sends message to certain user
	lobby.chat <- fmt.Sprintf(`%v%v<b title='%v' style='color: %v'>%v</b>`, tag, serverPrefix, game.Now(), color, msg)
}

func getArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	var args []string
	for start, n := 0, 0; n < len(raw); n++ {
		if raw[n] == 32 {
			args = append(args, raw[start:n])
			start = n + 1 //start at next letter
		} else if raw[n] == 34 {
			n++ //skip over opening quote
			for start = n; n < len(raw); n++ {
				if raw[n] == 34 { //go until ending quote
					break
				}
			}
			args = append(args, raw[start:n])
			if n+2 < len(raw) { //skip ending quote and space if possible
				n++ //n++ will be directly called again after this
				start = n + 1
			}
		} else if n == len(raw)-1 {
			args = append(args, raw[start:n+1])
		}
	}
	return args
}

func parseWords(rawWords *string) (*[]string, *string) {
	words := strings.Split(*rawWords, `,`)
	wordList := ``
	for n := range words {
		words[n] = strings.TrimSpace(words[n])
		wordList += fmt.Sprintf(`"%v"<br>`, words[n])
	}
	return &words, &wordList
}
