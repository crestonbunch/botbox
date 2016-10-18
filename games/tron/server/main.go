package main

import (
	"github.com/crestonbunch/botbox/common/game"
	"github.com/crestonbunch/botbox/games/tron"
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

	game.RunAuthenticatedServer(
		func(idList, secretList []string) (*game.GameServer, error) {
			return game.NewSynchronizedGameServer(
				tron.NewTwoPlayerTron(32, 32),
				idList,
				secretList,
			)
		},
	)

}
