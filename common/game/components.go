package game

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"time"
)

const StateLogFile = "./state.log"
const ResultLogFile = "./result.log"
const ConnectLogFile = "./connect.log"
const DisconnectLogFile = "./disconnect.log"

const ConnTimeout = 10 // sec
const MoveTimeout = 10 // sec

// A generic game server interface.
type GameServer struct {
	ConnectionManager
	ClientManager
	StateManager
	*GameRecorder
}

type GameClient interface {
	// Get a unique identifier for this client.
	Id() string
	// Get a websocket connection for this client
	Conn() *websocket.Conn
	// Return a watchdog for this client with a timeout on how long it can take
	// to make a move. Keeps clients from blocking forever.
	Watchdog() *Watchdog
}

// This component handles registering, disconnecting, scoring players, etc.
type ClientManager interface {
	// Register an agent to play from an http request, or return an error if there
	// was a problem authenticating the connection.
	Register(*websocket.Conn) (GameClient, error)
	// Get a list of the connected clients.
	Clients() []GameClient
}

// A manager that handles how actions should be received from clients. I.e.,
// real-time, synchronous, turn-based, etc.
type StateManager interface {
	// Given a list of game clients, spawn a goroutine and play the game by
	// sending/receiving messages according to how the game should progress.
	// Return a channel which will receive the game state every time it changes.
	Play([]GameClient, chan GameState, chan error)
}

type ConnectionManager interface {
	// Wait for clients to connect to the given URI path.
	// Provide a listener to register them with
	// the authentication manager and client managers, etc.
	// The listener should return a boolean whether the connection was accepted
	// or not.
	Wait(path string, listener func(conn *websocket.Conn) bool) error

	// Close all of the active connections.
	Close() error
}

func NewGameServer(
	connMan ConnectionManager,
	clientMan ClientManager,
	stateMan StateManager,
	recorder *GameRecorder,
) *GameServer {
	return &GameServer{
		connMan,
		clientMan,
		stateMan,
		recorder,
	}
}

func (s *GameServer) Start(path string) {
	defer s.Close()

	log.Println("Started server at path: " + path)
	listener := func(conn *websocket.Conn) bool {
		log.Println("Client connection received.")

		client, err := s.ClientManager.Register(conn)

		if err != nil {
			log.Println("Client rejected: " + err.Error())
			return false
		}

		log.Println("Client connected: authentication success.")
		s.GameRecorder.LogConnection(client)

		return true
	}

	log.Println("Waiting for connections.")

	err := s.ConnectionManager.Wait(path, listener)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("All clients connected.")

	// Play the game
	stateChan := make(chan GameState)
	errChan := make(chan error)
	go s.StateManager.Play(s.ClientManager.Clients(), stateChan, errChan)

	for {
		select {
		case err := <-errChan:
			switch err.(type) {
			case ClientError:
				log.Println("Client committed a sin: " + err.Error())
				s.GameRecorder.LogDisconnection(err.(ClientError).client)
			default:
				log.Println("Non-client related error occurred.")
			}
		case state := <-stateChan:
			// a state change has occurred
			s.GameRecorder.LogState(state)

			if state.Finished() {
				s.GameRecorder.LogResult(state)
				log.Printf("Result: %s\n", state.Result())
				log.Println("Game over.")
				return
			}
		}
	}
}

func (s *GameServer) Close() {
	// close the game recorder
	if err := s.GameRecorder.Close(); err != nil {
		log.Println("Error closing game recorder: " + err.Error())
	}
	// close client connections
	if err := s.ConnectionManager.Close(); err != nil {
		log.Println("Error: " + err.Error())
	}

	// This is the easiest way of killing the HTTP server
	os.Exit(0)
}

// A simple connection manager connects clients and requires them to all
// connect before moving forward. Sets a timeout so that clients must connect
// within the time frame or an error is returned to the game manager.
type SimpleConnectionManager struct {
	Connections    []*websocket.Conn
	NumConnections int
}

func NewSimpleConnectionManager(num int) *SimpleConnectionManager {
	return &SimpleConnectionManager{make([]*websocket.Conn, 0, num), num}
}

func (m *SimpleConnectionManager) Wait(
	path string,
	listener func(*websocket.Conn) bool,
) error {

	errChan := make(chan error)

	// handler to register connected clients
	http.Handle(path, websocket.Handler(func(conn *websocket.Conn) {
		defer func() {
			err := conn.Close()
			if err != nil {
				errChan <- err
			}
		}()

		accept := listener(conn)

		if accept {
			m.Connections = append(m.Connections, conn)
		} else {
			// reject this client, it was not accepted by the listener
			return
		}

		log.Printf("Connections: %d\\%d\n", len(m.Connections), m.NumConnections)

		for {
			// keep the connection alive while the agents play the game
		}
	}))

	watchdog := NewWatchdog(ConnTimeout)
	watchdog.Start(errors.New("Connection timeout."), errChan)

	// block until all clients have connected or TODO: add a timeout
	for len(m.Connections) < m.NumConnections {
		select {
		case err := <-errChan:
			// if there is an error, send it to the main go routine
			return err
		default:
			// don't block
			// TODO: measure a timeout
		}
	}

	watchdog.Stop()

	return nil
}

func (m *SimpleConnectionManager) Close() error {
	var result error = nil
	for _, c := range m.Connections {
		if err := c.Close(); err != nil {
			result = err
		}
	}
	return result
}

// An authenticated client manager will require secret keys passed in for each
// client id. If a client does not pass in a valid key, or passes in a
// duplicate key, then it will be rejected.
type AuthenticatedClientManager struct {
	constructor   func(id string, conn *websocket.Conn) GameClient
	clients       []GameClient
	clientIds     []string
	clientSecrets []string
}

// Create a new simple client manager. Give it a constructor to create clients
// from connections and also a list of client ids and secrets to expect.
func NewAuthenticatedClientManager(
	constructor func(id string, conn *websocket.Conn) GameClient,
	clientIds []string,
	clientSecrets []string,
) *AuthenticatedClientManager {
	return &AuthenticatedClientManager{
		constructor,
		make([]GameClient, 0, len(clientIds)),
		clientIds,
		clientSecrets,
	}
}

func (m *AuthenticatedClientManager) Register(conn *websocket.Conn) (GameClient, error) {
	secret := conn.Request().Header.Get("Authorization")

	if secret == "" {
		// no key sent by client
		return nil, errors.New("Secret is required.")
	}

	// check if secret is in the list
	valid := false
	var client GameClient = nil
	for i, k := range m.clientSecrets {
		if k == secret {
			valid = true
			id := m.clientIds[i]
			// remove used secrets
			m.clientSecrets = append(m.clientSecrets[:i], m.clientSecrets[i+1:]...)
			m.clientIds = append(m.clientIds[:i], m.clientIds[i+1:]...)

			client = m.constructor(id, conn)
			m.clients = append(m.clients, client)
			break
		}
	}

	if valid == false {
		return nil, errors.New("Invalid secret.")
	}

	return client, nil
}

func (m *AuthenticatedClientManager) Clients() []GameClient {
	return m.clients
}

// A simple client manager tracks client connections, but does not care at
// all about authentication. It may be bad to use this in production! It will
// take client ids from the User-Agent HTTP header.
type SimpleClientManager struct {
	constructor func(id string, conn *websocket.Conn) GameClient
	clients     []GameClient
}

func NewSimpleClientManager(
	constructor func(id string, conn *websocket.Conn) GameClient,
) *SimpleClientManager {
	return &SimpleClientManager{constructor, []GameClient{}}
}

func (m *SimpleClientManager) Register(conn *websocket.Conn) (GameClient, error) {
	id := conn.Request().Header.Get("User-Agent")

	client := m.constructor(id, conn)
	m.clients = append(m.clients, client)

	return client, nil
}

func (m *SimpleClientManager) Clients() []GameClient {
	return m.clients
}

// The game recorder records the game state every time it changes and whether
// a client connects as expected or disconnects unexpectedly. This allows
// the sandbox service to adequately punish clients which are not well-behaved,
// and send game results to the scoreboard service.
type GameRecorder struct {
	StateLog      *os.File
	ResultLog     *os.File
	ConnectLog    *os.File
	DisconnectLog *os.File
}

func NewGameRecorder(
	statePath,
	resultPath,
	connectPath,
	disconnectPath string,
) (*GameRecorder, error) {
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := os.FileMode(0600)
	stateLog, err := os.OpenFile(statePath, flags, perm)
	if err != nil {
		return nil, err
	}
	resultLog, err := os.OpenFile(resultPath, flags, perm)
	if err != nil {
		return nil, err
	}
	connectLog, err := os.OpenFile(connectPath, flags, perm)
	if err != nil {
		return nil, err
	}
	disconnectLog, err := os.OpenFile(disconnectPath, flags, perm)
	if err != nil {
		return nil, err
	}

	return &GameRecorder{stateLog, resultLog, connectLog, disconnectLog}, nil
}

func (r *GameRecorder) LogState(s GameState) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = r.StateLog.Write(append(b, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func (r *GameRecorder) LogResult(s GameState) error {
	b, err := json.Marshal(s.Result())
	if err != nil {
		return err
	}
	_, err = r.ResultLog.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameRecorder) LogConnection(c GameClient) error {
	_, err := r.ConnectLog.WriteString(c.Id() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (r *GameRecorder) LogDisconnection(c GameClient) error {
	_, err := r.DisconnectLog.WriteString(c.Id() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (r *GameRecorder) Close() error {
	if err := r.StateLog.Close(); err != nil {
		return err
	}
	if err := r.ResultLog.Close(); err != nil {
		return err
	}
	if err := r.DisconnectLog.Close(); err != nil {
		return err
	}
	if err := r.ConnectLog.Close(); err != nil {
		return err
	}
	return nil
}

type ClientMessage struct {
	Action string `json:"action"`
}

type ServerMessage struct {
	Player  int         `json:"player"`
	Actions interface{} `json:"actions"`
	State   interface{} `json:"state"`
}

// This is an error that is associated with a client so that we can adequately
// punish clients who do not have good behavior.
type ClientError struct {
	err    error
	client GameClient
}

func (e ClientError) Error() string {
	return e.err.Error()
}

// A watchdog is a simple tool that will return an error if it is not reset
// by the time the timeout is up.
type Watchdog struct {
	timeout time.Duration
	timer   *time.Timer
}

func NewWatchdog(timeout int) *Watchdog {
	return &Watchdog{timeout: time.Duration(timeout) * time.Second}
}

// Start the watchdog on a separate goroutine. Will send an err to the
// given channel if it runs out of time before it is stopped or reset.
func (w *Watchdog) Start(err error, errChan chan error) {
	go func() {
		w.timer = time.AfterFunc(w.timeout, func() {
			log.Println("Watchdog timed out.")
			errChan <- err
		})
	}()
}

// Reset the watchdog timer to the initial value.
func (w *Watchdog) Reset() {
	if w.timer != nil {
		if !w.timer.Stop() {
			<-w.timer.C
		}
		w.timer.Reset(w.timeout)
	}
}

// Stop the watchdog from sending an error when the timeout is reached.
func (w *Watchdog) Stop() {
	if w.timer != nil {
		w.timer.Stop()
	}
}
