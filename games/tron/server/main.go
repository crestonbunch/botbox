package tron

import (
	"botbox"
	"golang.org/x/net/websocket"
	"net/http"
)

// Setup the tron server to listen to clients.
func main() {
	server := botbox.NewSynchronizedGameServer(NewTwoPlayerTron(32, 32), 2)

	go server.Start("/")

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
