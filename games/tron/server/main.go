package main

import (
	"botbox/common/server"
	"botbox/games/tron"
	"fmt"
	"net/http"
	"os"
)

// Setup the tron server to listen to clients.
// If command line arguments are provided, they are assumed to be client keys
// to expect, and authorization is required. If no client keys are provided,
// then authorization is not necessary.
func main() {
	keys := os.Args[1:]

	if len(keys) > 0 {
		fmt.Println("Requiring keys:", keys)
		s := server.NewAuthenticatedSynchronizedGameServer(tron.NewTwoPlayerTron(32, 32), keys)
		go s.Start("/")
	} else {
		s := server.NewSynchronizedGameServer(tron.NewTwoPlayerTron(32, 32), 2)
		go s.Start("/")
	}

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
