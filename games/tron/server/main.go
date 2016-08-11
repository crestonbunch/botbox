package main

import (
	"botbox/common/server"
	"botbox/games/tron"
	"net/http"
)

// Setup the tron server to listen to clients.
func main() {
	s := server.NewSynchronizedGameServer(tron.NewTwoPlayerTron(32, 32), 2)

	go s.Start("/")

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
