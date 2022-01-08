package tests

import (
	"Palette/lobby"
	"Palette/lobby/user"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

//ManagerTest is a very simple test that tests the functionality of the lobby database.
//I know that Go has built in unit tests but for a simple test like this, it doesn't matter.
func Manager(nLobbies int, manager *lobby.Manager, MAX_USER_TIMEOUT time.Duration) {
	start := time.Now()
	for i := 1; i <= nLobbies; i++ { //create n users and n lobbies with those users as the ID
		ID := fmt.Sprint(i)
		lobbyKey := ID
		host := user.New(ID)
		Lobby := lobby.New(lobbyKey, lobbyKey, MAX_USER_TIMEOUT, host) //create a new lobby
		manager.AddLobby(Lobby)                                        //add that lobby to the manager's list of lobbies

		//add user tests
		newHost := user.New(fmt.Sprint(i * 10)) //create a new user to be the new host
		nonUniqueNamedUser := user.New(ID)      //create a user with an adjusted name
		Lobby.AddUser(newHost)
		Lobby.AddUser(nonUniqueNamedUser)

		//remove user tests
		maxUserTimeoutAgo := time.Now().Add(-MAX_USER_TIMEOUT)
		host.SetTimeDisconnect(maxUserTimeoutAgo)                                //remove host, `lobby.userManager()` should remove the user, new host should be i*10
		newHost.SetTimeDisconnect(maxUserTimeoutAgo.Add(time.Second))            //remove new host
		nonUniqueNamedUser.SetTimeDisconnect(maxUserTimeoutAgo.Add(time.Second)) //remove name adjusted user, lobby should then be deleted as the lobby is empty
		if i == 1 {
			log.Println(nonUniqueNamedUser.Name()) //should be `ID-1` as it will be adjusted by `lobby.AddUser()`
			log.Printf(`Continue?`)
			fmt.Scanln() //block the first time around to see if the values are correct
		}
	}
	for range time.NewTicker(time.Second).C {
		if runtime.NumGoroutine() == 2 { //block and ensure that only two threads remain: [main, manager.cleanup()], if so all tests passed
			log.Println(`Sucessful test. Elapsed time:`, time.Since(start))
			os.Exit(0)
		}
		log.Println(runtime.NumGoroutine(), `active goroutines`)
	}
}
