package game

import (
	"flag"
	"github.com/crestonbunch/botbox/services/sandbox"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"strings"
)

// Setup a server to listen to clients.
// To start the server you must provide a list of ids and secrets. When
// a client connects with a valid secret, it will be automatically assigned the
// corresponding id. To give a list like this via the command line, call
// go run main.go --ids "1 2" --secrets "s1 s2"
// Otherwise, in a Docker sandbox you can set the environment variables
// BOTBOX_IDS and BOTBOX_SECRETS as space-separated lists of ids and secrets.
// The secrets are necessary to prevent malicious agents from trying to connect
// as two separate agents. Each client that connects should be given a secret
// but not told what any other secrets are.
// Pass in a constructor function that will build the GameServer from a list
// of client ids and secrets to expect.
func RunAuthenticatedServer(
	constructor func(ids, secrets []string) (websocket.Handler, error),
) {

	idList := []string{}
	secretList := []string{}
	if len(os.Args) > 1 {
		ids := flag.String("ids", "", "A space-delimited list of client ids.")
		secrets := flag.String("secrets", "", "A space-delimited list of client secrets.")

		flag.Parse()

		idList = strings.Split(*ids, " ")
		secretList = strings.Split(*secrets, " ")

	} else {
		ids, idsExist := os.LookupEnv(sandbox.ServerIdsEnvVar)
		secrets, secretsExist := os.LookupEnv(sandbox.ServerSecretEnvVar)
		if idsExist && secretsExist {
			idList = strings.Split(ids, sandbox.EnvListSep)
			secretList = strings.Split(secrets, sandbox.EnvListSep)
		}
	}

	if len(idList) == 0 || len(secretList) == 0 {
		panic("Must have ids and secrets for clients!")
	}

	if len(idList) != len(secretList) {
		panic("Ids and secret must be equal in length!")
	}

	log.Printf("Clients: %s\n", idList)
	log.Printf("Secrets: %s\n", secretList)

	handler, err := constructor(idList, secretList)
	if err != nil {
		panic(err)
	}
	http.Handle("/", handler)

	err = http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
