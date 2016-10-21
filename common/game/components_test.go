package game

import (
	"golang.org/x/net/websocket"
	"io/ioutil"
	"math"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
	"time"
)

func setupTestServer(handler websocket.Handler) (string, *httptest.Server) {
	ts := httptest.NewServer(handler)

	// change URL scheme from http:// to ws:// on the mock server
	url, _ := url.Parse(ts.URL)
	url.Scheme = "ws"

	return url.String(), ts
}

func TestSimpleConnectionManager(t *testing.T) {
	m := NewSimpleConnectionManager()
	defer m.Close()
	connChan := make(chan *websocket.Conn)
	errChan := make(chan error)

	handler := m.Handler(connChan, errChan)
	url, ts := setupTestServer(handler)
	defer ts.Close()
	origin := "http://localhost/"

	_, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Error(err)
	}

	select {
	case <-connChan:
	case err := <-errChan:
		t.Error(err)
	default:
		t.Error("No connection or error received.")
	}

	if len(m.Connections) != 1 {
		t.Error("SimpleConnectionManager does not have 1 connection.")
	}
}

func TestAuthenticatedClientManager(t *testing.T) {
	connChan := make(chan *websocket.Conn)
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	m := NewAuthenticatedClientManager(
		func(id string, conn *websocket.Conn) GameClient {
			return nil
		},
		ids,
		secrets,
		ConnTimeout,
	)

	wg := m.Register(connChan)

	url, ts := setupTestServer(func(conn *websocket.Conn) {
		connChan <- conn
	})
	defer ts.Close()
	origin := "http://localhost/"

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
	}

	wg.Wait()

	if !m.Ready() {
		t.Error("Client manager timed out waiting for connections.")
	}
}

func TestAuthenticatedTimeout(t *testing.T) {
	connChan := make(chan *websocket.Conn)
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	m := NewAuthenticatedClientManager(
		func(id string, conn *websocket.Conn) GameClient {
			return nil
		},
		ids,
		secrets,
		time.Millisecond,
	)

	start := time.Now()
	wg := m.Register(connChan)

	url, ts := setupTestServer(func(conn *websocket.Conn) {
		connChan <- conn
	})
	defer ts.Close()
	origin := "http://localhost/"

	for i := 0; i < 1; i++ {
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
	}

	wg.Wait()
	duration := time.Since(start)

	if m.Ready() {
		t.Error("Client manager should have timed out waiting for connections.")
	}

	if len(m.Clients()) != 1 {
		t.Error("Client manager does not have 1 connected client.")
	}

	// TODO: this is a janky way of testing timeouts
	if math.Abs(float64(duration-time.Millisecond)) < 0.1 {
		t.Error("Client manager did not timeout in 1 ms")
	}
}

func TestAuthenticatedNoSecret(t *testing.T) {
	connChan := make(chan *websocket.Conn)
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	m := NewAuthenticatedClientManager(
		func(id string, conn *websocket.Conn) GameClient {
			return nil
		},
		ids,
		secrets,
		time.Millisecond,
	)

	start := time.Now()
	wg := m.Register(connChan)

	url, ts := setupTestServer(func(conn *websocket.Conn) {
		connChan <- conn
	})
	defer ts.Close()
	origin := "http://localhost/"

	for i := 0; i < 1; i++ {
		config, err := websocket.NewConfig(url, origin)
		if err != nil {
			t.Error(err)
		}
		//config.Header.Add("Authorization", secrets[i])
		conn, err := websocket.DialConfig(config)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
	}

	wg.Wait()
	duration := time.Since(start)

	if m.Ready() {
		t.Error("Client manager should have timed out waiting for connections.")
	}

	if len(m.Clients()) > 0 {
		t.Error("Client manager does not have any connected clients.")
	}

	// TODO: this is a janky way of testing timeouts
	if math.Abs(float64(duration-time.Millisecond)) < 0.1 {
		t.Error("Client manager did not timeout in 1 ms")
	}
}

func TestAuthenticatedBadAgents(t *testing.T) {
	connChan := make(chan *websocket.Conn)
	ids := []string{"id1", "id2"}
	secrets := []string{"secret1", "secret2"}
	m := NewAuthenticatedClientManager(
		func(id string, conn *websocket.Conn) GameClient {
			return nil
		},
		ids,
		secrets,
		ConnTimeout,
	)

	m.Register(connChan)

	url, ts := setupTestServer(func(conn *websocket.Conn) {
		connChan <- conn
	})
	defer ts.Close()
	origin := "http://localhost/"

	for i := 0; i < 2; i++ {
		config, err := websocket.NewConfig(url, origin)
		if err != nil {
			t.Error(err)
		}
		config.Header.Add("Authorization", "blah")
		conn, err := websocket.DialConfig(config)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
	}

	if len(m.Clients()) != 0 {
		t.Error("Client manager has more than 0 connections!")
	}
}

func TestSimpleGameRecorder(t *testing.T) {

	dir := os.TempDir()
	r, err := NewSimpleGameRecorder(dir)
	if err != nil {
		t.Error(err)
	}
	defer r.Close()

	testState := &mockState{[]int{0, 0}}
	r.LogState(testState)
	testState = &mockState{[]int{1, 0}}
	r.LogState(testState)

	if f, err := os.Open(path.Join(dir, StateLogFile)); err != nil {
		t.Error(err)
	} else {
		defer f.Close()
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		if string(contents) != "{\"players\":[0,0]}\n{\"players\":[1,0]}\n" {
			t.Error("GameRecorder did not record correct state")
		}
	}

	testState = &mockState{[]int{12, 4}}
	r.LogResult(testState)

	if f, err := os.Open(path.Join(dir, ResultLogFile)); err != nil {
		t.Error(err)
	} else {
		defer f.Close()
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		if string(contents) != "[1,-1]" {
			t.Error("GameRecorder did not record correct result.")
		}
	}

	testClient := &SynchronizedGameClient{id: "123abc"}
	r.LogConnection(testClient)
	testClient = &SynchronizedGameClient{id: "456def"}
	r.LogConnection(testClient)

	if f, err := os.Open(path.Join(dir, ConnectLogFile)); err != nil {
		t.Error(err)
	} else {
		defer f.Close()
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		if string(contents) != "123abc\n456def\n" {
			t.Error("GameRecorder did not record correct connections.")
		}
	}

	testClient = &SynchronizedGameClient{id: "123abc"}
	r.LogDisconnection(testClient)
	testClient = &SynchronizedGameClient{id: "456def"}
	r.LogDisconnection(testClient)

	if f, err := os.Open(path.Join(dir, DisconnectLogFile)); err != nil {
		t.Error(err)
	} else {
		defer f.Close()
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		if string(contents) != "123abc\n456def\n" {
			t.Error("GameRecorder did not record correct disconnections.")
		}
	}

}

func mockTwoPlayerGame() *mockState {
	return &mockState{[]int{0, 0}}
}

type mockGameRecorder struct{}

func (r *mockGameRecorder) LogState(s GameState) error {
	return nil
}

func (r *mockGameRecorder) LogResult(s GameState) error {
	return nil
}

func (r *mockGameRecorder) LogConnection(c GameClient) error {
	return nil
}

func (r *mockGameRecorder) LogDisconnection(c GameClient) error {
	return nil
}

func (r *mockGameRecorder) Close() error {
	return nil
}

type mockState struct {
	Players []int `json:"players"`
}

func (s *mockState) Actions(p int) interface{} {
	return []string{"1", "2", "3"}
}

func (s *mockState) Do(p int, a string) {
	if a == "1" {
		s.Players[p] += 1
	} else if a == "2" {
		s.Players[p] += 2
	} else if a == "3" {
		s.Players[p] += 3
	}
}

func (s *mockState) View(p int) interface{} {
	return s
}

func (s *mockState) Finished() bool {
	for _, p := range s.Players {
		if p >= 10 {
			return true
		}
	}

	return false
}

func (s *mockState) Result() []int {
	result := []int{}
	max := -1
	for _, p := range s.Players {
		if p > max {
			max = p
		}
	}
	count := 0
	for _, p := range s.Players {
		if p == max {
			result = append(result, ResultWin)
			count += 1
		} else {
			result = append(result, ResultLoss)
		}
	}
	if count > 1 {
		for i, p := range s.Players {
			if p == max {
				result[i] = ResultTie
			}
		}
	}

	return result
}
