package main

import (
	"github.com/crestonbunch/botbox/common/game"
	"github.com/crestonbunch/botbox/games/tron"
	"golang.org/x/net/websocket"
)

// Setup the tron server to listen to clients.
// To start the server you must provide a list of ids and secrets. When
// a client connects with a valid secret, it will be automatically assigned the
// corresponding id. To give a list like this via the command line, call
// go run main.go --ids "1 2" --secrets "s1 s2"
// Otherwise, in a Docker sandbox you can set the environment variables
// BOTBOX_IDS and BOTBOX_SECRETS as space-separated lists of ids and secrets.
// The secrets are necessary to prevent malicious agents from trying to connect
// as two separate agents. Each client that connects should be given a secret
// but not told what any other secrets are.
func main() {

	exitChan := make(chan bool)

	go func() {
		game.RunAuthenticatedServer(
			func(idList, secretList []string) (websocket.Handler, error) {
				writer, err := game.NewSimpleGameRecorder("./")
				if err != nil {
					return nil, err
				}
				stateMan := game.NewSynchronizedStateManager(
					tron.NewTwoPlayerTron(32, 32), game.MoveTimeout,
				)

				return game.GameHandler(
					exitChan,
					game.NewSimpleConnectionManager(),
					game.NewAuthenticatedClientManager(
						stateMan.NewClient, idList, secretList, game.ConnTimeout,
					),
					stateMan,
					writer,
				), nil
			},
		)
	}()

	<-exitChan
}
