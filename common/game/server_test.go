package game

import (
	"github.com/crestonbunch/botbox/services/sandbox"
	"golang.org/x/net/websocket"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthenticateHandler(t *testing.T) {
	exitChan := make(chan bool)
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	err := os.Setenv(
		sandbox.ServerIdsEnvVar, strings.Join(ids, sandbox.EnvListSep),
	)
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv(sandbox.ServerIdsEnvVar)
	err = os.Setenv(
		sandbox.ServerSecretEnvVar, strings.Join(secrets, sandbox.EnvListSep),
	)
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv(sandbox.ServerSecretEnvVar)

	constructor := func(ids, secrets []string) (websocket.Handler, error) {
		connMan := NewSimpleConnectionManager()
		state := &mockState{[]int{0, 0}}
		stateMan := NewSynchronizedStateManager(state, time.Second)
		clientMan := NewAuthenticatedClientManager(
			stateMan.NewClient,
			ids,
			secrets,
			time.Second,
		)
		recorder := &mockGameRecorder{}

		return GameHandler(
			exitChan,
			connMan,
			clientMan,
			stateMan,
			recorder,
		), nil
	}
	handler, err := AuthenticateHandler(constructor)
	if err != nil {
		t.Error(err)
	}

	url, ts := setupTestServer(handler)
	defer ts.Close()
	origin := "http://localhost/"
	conns := []*websocket.Conn{}
	for i := 0; i < 2; i++ {
		config, err := websocket.NewConfig(url, origin)
		if err != nil {
			t.Error(err)
		}
		config.Header.Add("Authorization", secrets[i])
		conn, err := websocket.DialConfig(config)
		if err != nil {
			t.Error(err)
		}
		conns = append(conns, conn)
	}

	// Simulate two players
	for i := 0; i < 4; i++ {
		go func() {
			var msg ServerMessage
			err := websocket.JSON.Receive(conns[0], &msg)
			if err != nil {
				t.Error(err)
			}
			broadcast := ClientMessage{Action: "1"}
			err = websocket.JSON.Send(conns[0], &broadcast)
			if err != nil {
				t.Error(err)
			}
		}()

		go func() {
			var msg ServerMessage
			err := websocket.JSON.Receive(conns[1], &msg)
			if err != nil {
				t.Error(err)
			}
			broadcast := ClientMessage{Action: "3"}
			err = websocket.JSON.Send(conns[1], &broadcast)
			if err != nil {
				t.Error(err)
			}
		}()

	}

	<-exitChan
}

func TestRequireSecretsIds(t *testing.T) {
	exitChan := make(chan bool)
	constructor := func(ids, secrets []string) (websocket.Handler, error) {
		connMan := NewSimpleConnectionManager()
		state := &mockState{[]int{0, 0}}
		stateMan := NewSynchronizedStateManager(state, time.Second)
		clientMan := NewAuthenticatedClientManager(
			stateMan.NewClient,
			ids,
			secrets,
			time.Second,
		)
		recorder := &mockGameRecorder{}

		return GameHandler(
			exitChan,
			connMan,
			clientMan,
			stateMan,
			recorder,
		), nil
	}

	_, err := AuthenticateHandler(constructor)
	if err == nil {
		t.Error("Secrets and ids are not required!")
	}
}

func TestRequireEqualSecretsIds(t *testing.T) {
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2", "secret3"}
	err := os.Setenv(
		sandbox.ServerIdsEnvVar, strings.Join(ids, sandbox.EnvListSep),
	)
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv(sandbox.ServerIdsEnvVar)
	err = os.Setenv(
		sandbox.ServerSecretEnvVar, strings.Join(secrets, sandbox.EnvListSep),
	)
	if err != nil {
		t.Error(err)
	}
	defer os.Unsetenv(sandbox.ServerSecretEnvVar)

	exitChan := make(chan bool)
	constructor := func(ids, secrets []string) (websocket.Handler, error) {
		connMan := NewSimpleConnectionManager()
		state := &mockState{[]int{0, 0}}
		stateMan := NewSynchronizedStateManager(state, time.Second)
		clientMan := NewAuthenticatedClientManager(
			stateMan.NewClient,
			ids,
			secrets,
			time.Second,
		)
		recorder := &mockGameRecorder{}

		return GameHandler(
			exitChan,
			connMan,
			clientMan,
			stateMan,
			recorder,
		), nil
	}

	_, err = AuthenticateHandler(constructor)
	if err == nil {
		t.Error("Secrets and ids are not required!")
	}
}
