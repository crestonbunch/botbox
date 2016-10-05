package main

import (
	"github.com/crestonbunch/botbox/common/server"
	"github.com/crestonbunch/botbox/games/tron"
	"github.com/crestonbunch/botbox/services/sandbox"
	"log"
	"net/http"
	"os"
	"strings"
)

// Setup the tron server to listen to clients.
// If command line arguments are provided, they are assumed to be client keys
// to expect, and authorization is required. If no client keys are provided,
// then authorization is not necessary.
func main() {

	keys := []string{}
	if env, exist := os.LookupEnv(sandbox.ServerSecretEnvVar); exist {
		keys = strings.Split(env, sandbox.EnvListSep)
	} else {
		keys = os.Args[1:]
	}

	if len(keys) > 0 {
		log.Println("Requiring keys:", keys)
		s := server.NewAuthenticatedSynchronizedGameServer(tron.NewTwoPlayerTron(32, 32), keys)
		go s.Start("/")
	} else {
		log.Println("Not requiring keys")
		s := server.NewSynchronizedGameServer(tron.NewTwoPlayerTron(32, 32), 2)
		go s.Start("/")
	}

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
