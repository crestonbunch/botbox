package botbox

import (
	"encoding/json"
	"golang.org/x/net/websocket"
	"net/http"
	"testing"
)

type dummySynchronizedState struct {
	P1 int `json:"p1"`
	P2 int `json:"p2"`
}

func (s *dummySynchronizedState) Actions(p int) interface{} {
	return []string{"1", "2"}
}

func (s *dummySynchronizedState) Do(p int, a interface{}) {
	switch p {
	case 0:
		switch a.(string) {
		case "1":
			s.P1 += 1
		case "2":
			s.P1 += 2
		}
	case 1:
		switch a.(string) {
		case "1":
			s.P2 += 1
		case "2":
			s.P2 += 2
		}
	}
}

func (s *dummySynchronizedState) View(p int) []byte {
	result, _ := json.Marshal(s)
	return result
}

func (s *dummySynchronizedState) Finished() bool {
	return s.P1 == 10 && s.P2 == 10
}

func runDummyClient(origin, url string) {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < 10; i++ {
		var turn TurnMessage
		err := websocket.JSON.Receive(ws, &turn)
		// listen for turn signal
		if err != nil {
			panic(err.Error())
		}
		// make an action
		response := GameMessage{[]byte("do"), "1"}
		websocket.JSON.Send(ws, response)
	}
}

func TestSynchronizedGameServer(t *testing.T) {
	state := &dummySynchronizedState{}
	server := NewSynchronizedGameServer(state, 2)

	go server.Start("/")

	go func() {
		err := http.ListenAndServe(":12345", nil)
		if err != nil {
			panic(err.Error())
		}
	}()

	origin := "http://localhost/"
	url := "ws://localhost:12345/"
	// simulate two clients
	go runDummyClient(origin, url)
	go runDummyClient(origin, url)

	for !state.Finished() {
		// block until game is over
	}

	if state.P1 != 10 {
		t.Error("Player 1 did not reach 10 is at", state.P1)
	}

	if state.P2 != 10 {
		t.Error("Player 2 did not reach 10 is at", state.P2)
	}
}
