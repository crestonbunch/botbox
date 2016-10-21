package game

import (
	"errors"
	"flag"
	"github.com/crestonbunch/botbox/services/sandbox"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"strings"
)

var ids string
var secrets string

func SetupFlags() {
	flag.StringVar(&ids, "ids", "", "A space-delimited list of client ids.")
	flag.StringVar(&secrets, "secrets", "", "A space-delimited list of client secrets.")
	flag.Parse()
}

// Searches first for command line arguments "ids" and "secrets", and then
// checks environment variables. Returns an error if they were not found or
// they were not the same length.
func FindIdsAndSecrets() ([]string, []string, error) {

	idList := []string{}
	secretList := []string{}
	if ids != "" && secrets != "" {
		idList = strings.Split(ids, " ")
		secretList = strings.Split(secrets, " ")
	} else {
		ids, idsExist := os.LookupEnv(sandbox.ServerIdsEnvVar)
		secrets, secretsExist := os.LookupEnv(sandbox.ServerSecretEnvVar)
		if idsExist && secretsExist {
			idList = strings.Split(ids, sandbox.EnvListSep)
			secretList = strings.Split(secrets, sandbox.EnvListSep)
		}
	}

	if len(idList) == 0 || len(secretList) == 0 {
		return nil, nil, errors.New("Must have ids and secrets for clients!")
	}

	if len(idList) != len(secretList) {
		return nil, nil, errors.New("Must have equal number ids and secrets!")
	}

	return idList, secretList, nil

}

// Given a constructor that creates a websocket handler, wrap it with
// FindIdsAndSecrets() to authenticate from the command line or environment
// variables.
func AuthenticateHandler(
	constructor func(ids, secrets []string) (websocket.Handler, error),
) (websocket.Handler, error) {
	idList, secretList, err := FindIdsAndSecrets()
	if err != nil {
		return nil, err
	}
	log.Println(idList)
	log.Println(secretList)
	return constructor(idList, secretList)
}

// Setup a server to listen to clients.
// To start the server you must provide a list of ids and secrets. When
// a client connects with a valid secret, it will be automatically assigned the
// corresponding id. To give a list like this via the command line, call
// go run main.go --ids "1 2" --secrets "s1 s2"
// Otherwise, in a Docker sandbox you can set the environment variables
// BOTBOX_IDS and BOTBOX_SECRETS as space-separated lists of ids and secrets.
// The secrets are necessary to prevent malicious scripts from trying to connect
// as two separate agents.
// Pass in a constructor function that will build the GameHandler from a list
// of client ids and secrets to expect.
func RunAuthenticatedServer(
	constructor func(ids, secrets []string) (websocket.Handler, error),
) {
	SetupFlags()

	handler, err := AuthenticateHandler(constructor)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", handler)

	err = http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}

}
