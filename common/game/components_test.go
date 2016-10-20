package game

import (
	"golang.org/x/net/websocket"
	"math"
	"net/http/httptest"
	"net/url"
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
