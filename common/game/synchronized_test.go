package game

import (
	"golang.org/x/net/websocket"
	"math"
	"testing"
	"time"
)

func TestSynchronizedStateManager(t *testing.T) {
	state := &mockState{[]int{0, 0}}
	stateChan := make(chan GameState)
	errChan := make(chan error)
	stateMan := NewSynchronizedStateManager(state, time.Millisecond)

	url, ts := setupTestServer(func(conn *websocket.Conn) {
		for {
		}
	})
	defer ts.Close()
	origin := "http://localhost/"

	conns := []*websocket.Conn{}
	for i := 0; i < 2; i++ {
		conn, err := websocket.Dial(url, "", origin)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()

		conns = append(conns, conn)
	}

	clients := []GameClient{
		stateMan.NewClient("1", conns[0]),
		stateMan.NewClient("2", conns[1]),
	}

	wg := stateMan.Play(clients, stateChan, errChan)

	go func() {
		for i := 0; i < 4; i++ {
			// make move
			<-clients[0].Send()
			clients[0].Receive() <- ClientMessage{"1"}
			<-clients[1].Send()
			clients[1].Receive() <- ClientMessage{"3"}
			select {
			case <-stateChan:
			case err := <-errChan:
				t.Error(err)
			}
		}
	}()

	wg.Wait()

	if !state.Finished() {
		t.Error("Game did not finish!")
	}

	result := state.Result()

	if result[0] != ResultLoss {
		t.Error("Player 1 did not lose")
	}
	if result[1] != ResultWin {
		t.Error("Player 2 did not win")
	}
}

func TestSynchronizedGameHandler(t *testing.T) {
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}

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
	exitChan := make(chan bool)

	handler := GameHandler(
		exitChan,
		connMan,
		clientMan,
		stateMan,
		recorder,
	)

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
		defer conn.Close()
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

	if !state.Finished() {
		t.Error("Game did not finish!")
	}

	result := state.Result()

	if result[0] != ResultLoss {
		t.Error("Player 1 did not lose")
	}
	if result[1] != ResultWin {
		t.Error("Player 2 did not win")
	}
	if state.Players[0] != 4 {
		t.Error("Player 1 score is not 4")
	}
	if state.Players[1] != 12 {
		t.Error("Player 2 score is not 12")
	}

}

func TestSynchronizedBadAgentConnectionTimeout(t *testing.T) {
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}

	connMan := NewSimpleConnectionManager()
	state := &mockState{[]int{0, 0}}
	stateMan := NewSynchronizedStateManager(state, time.Second)
	clientMan := NewAuthenticatedClientManager(
		stateMan.NewClient,
		ids,
		secrets,
		2*time.Millisecond,
	)
	recorder := &mockGameRecorder{}
	exitChan := make(chan bool)

	handler := GameHandler(
		exitChan,
		connMan,
		clientMan,
		stateMan,
		recorder,
	)

	sendSecrets := []string{"secret1", "badsecret"}

	start := time.Now()
	url, ts := setupTestServer(handler)
	defer ts.Close()
	origin := "http://localhost/"
	conns := []*websocket.Conn{}
	for i := 0; i < 2; i++ {
		config, err := websocket.NewConfig(url, origin)
		if err != nil {
			t.Error(err)
		}
		config.Header.Add("Authorization", sendSecrets[i])
		conn, err := websocket.DialConfig(config)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		conns = append(conns, conn)
	}

	<-exitChan

	duration := time.Since(start)

	if len(connMan.Connections) != 2 {
		t.Error("Connection manager does not have 2 connection.")
	}

	if len(clientMan.Clients()) != 1 {
		t.Error("Client manager does not have 1 client.")
	}

	// TODO: this is a janky way of testing timeouts
	if math.Abs(float64(duration-2*time.Millisecond)) < 0.1 {
		t.Error("Client manager did not timeout in 1 ms")
	}
}

func TestSynchronizedAttemptedCheater(t *testing.T) {
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}

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
	exitChan := make(chan bool)

	handler := GameHandler(
		exitChan,
		connMan,
		clientMan,
		stateMan,
		recorder,
	)

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
		defer conn.Close()
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
			// Player 1 tries to cheat by sending 10 actions in a row, but it should
			// not let him get away with that
			for i := 0; i < 10; i++ {
				go func() {
					broadcast := ClientMessage{Action: "1"}
					err = websocket.JSON.Send(conns[0], &broadcast)
					if err != nil {
						t.Error(err)
					}
				}()
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

	if !state.Finished() {
		t.Error("Game did not finish!")
	}

	result := state.Result()

	if result[0] != ResultLoss {
		t.Error("Player 1 did not lose")
	}
	if result[1] != ResultWin {
		t.Error("Player 2 did not win")
	}
	if state.Players[0] != 4 {
		t.Error("Player 1 score is not 4")
	}
	if state.Players[1] != 12 {
		t.Error("Player 2 score is not 12")
	}

}

func TestSynchronizedUselessAgent(t *testing.T) {
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}

	connMan := NewSimpleConnectionManager()
	state := &mockState{[]int{0, 0}}
	stateMan := NewSynchronizedStateManager(state, 2*time.Millisecond)
	clientMan := NewAuthenticatedClientManager(
		stateMan.NewClient,
		ids,
		secrets,
		time.Second,
	)
	recorder := &mockGameRecorder{}
	exitChan := make(chan bool)

	handler := GameHandler(
		exitChan,
		connMan,
		clientMan,
		stateMan,
		recorder,
	)

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
		defer conn.Close()
		conns = append(conns, conn)
	}

	// Simulate two players
	for i := 0; i < 4; i++ {
		go func() {
			// Player 1 just forgets to do anything at all!
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

	if !state.Finished() {
		t.Error("Game did not finish!")
	}

	result := state.Result()

	if result[0] != ResultLoss {
		t.Error("Player 1 did not lose")
	}
	if result[1] != ResultWin {
		t.Error("Player 2 did not win")
	}
	if state.Players[0] != 0 {
		t.Error("Player 1 score is not 0")
	}
	if state.Players[1] != 12 {
		t.Error("Player 2 score is not 12")
	}

}
