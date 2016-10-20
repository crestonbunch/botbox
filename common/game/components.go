package game

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"sync"
	"time"
)

const StateLogFile = "./state.log"
const ResultLogFile = "./result.log"
const ConnectLogFile = "./connect.log"
const DisconnectLogFile = "./disconnect.log"

const ConnTimeout = 10 * time.Second
const MoveTimeout = 10 * time.Second

type GameClient interface {
	// Get a unique identifier for this client.
	Id() string
	// Get a websocket connection for this client
	Conn() *websocket.Conn
	// Return a watchdog for this client with a timeout on how long it can take
	// to make a move. Keeps clients from blocking forever.
	Watchdog() *Watchdog

	// Send and receive channels for communicating with the client
	Send() chan ServerMessage
	Receive() chan ClientMessage
	Error() chan ClientError
}

type ConnectionManager interface {
	// Create a websocket handler that will send connections along the given
	// channel. This channel can be given to a listener which will do something
	// with the connections. E.g., a client manager. The last channel is
	// the exit channel, which should terminate the handler when it recieves a
	// value.
	Handler(chan *websocket.Conn, chan error) websocket.Handler

	// Close all of the active connections.
	Close()
}

// This component handles authenticating and managing clients.
type ClientManager interface {
	// Listen to a channel that receives websocket connections and authenticate
	// connections. If they are valid, allow them to remain connected other wise
	// close them immediately. Return a waitgroup that finished when all clients
	// are connected or a timeout occurs.
	Register(chan *websocket.Conn) *sync.WaitGroup
	// Return whether or not the client manager has all of the expected clients
	// connected.
	Ready() bool
	// Get a list of the connected clients.
	Clients() []GameClient
}

// A manager that handles how actions should be received from clients. I.e.,
// real-time, synchronous, turn-based, etc.
type StateManager interface {
	// Given a list of game clients, spawn a goroutine and play the game by
	// sending/receiving messages according to how the game should progress.
	// Return a channel which will receive the game state every time it changes.
	// Return a waitgroup that finished when the game is over or a timeout
	// occurs.
	Play([]GameClient, chan GameState, chan error) *sync.WaitGroup
}

type GameRecorder interface {
	LogState(GameState) error
	LogResult(GameState) error
	LogConnection(GameClient) error
	LogDisconnection(GameClient) error
	Close() error
}

// Start the game components and return a websocket handler that can be used
// to start or mock and HTTP server and receive requests. Must be given a
// connection manager, client manager, state manager, and game recorder. The
// first channel parameter will be passed a value when the game is over, that
// can be used to stop the HTTP server.
func GameHandler(
	exitChan chan bool,
	connMan ConnectionManager,
	clientMan ClientManager,
	stateMan StateManager,
	record GameRecorder,
) websocket.Handler {

	abortChan := make(chan bool)
	errChan := make(chan error)
	connChan := make(chan *websocket.Conn)
	stateChan := make(chan GameState)
	handler := connMan.Handler(connChan, errChan)

	go func() {
		log.Println("Waiting for connections.")
		// Bind the client manager to the connection channel so that each time a
		// connection is made by the connection manager, the client manager will
		// inspect it and accept / reject the client and close the connection if
		// necessary
		wgReg := clientMan.Register(connChan)
		// wait until all clients are connected or a timeout occurs
		wgReg.Wait()

		for _, c := range clientMan.Clients() {
			// Log clients that successfully connected.
			record.LogConnection(c)
			Listen(c)
		}

		if !clientMan.Ready() {
			// The client manager timed out before all clients connected
			log.Println("Client manager timeout. Exiting.")
			abortChan <- true
			return
		}

		log.Println("All clients connected.")
		// If all clients are connected, begin playing the game by sending the
		// request to the state manager to play.
		wgPlay := stateMan.Play(clientMan.Clients(), stateChan, errChan)
		// wait until the game is over or a timeout occurs
		wgPlay.Wait()
	}()

	go func() {
		defer connMan.Close()
		defer record.Close()
		defer func() {
			exitChan <- true
		}()
		for {
			select {
			case err := <-errChan:
				switch err.(type) {
				case ClientError:
					log.Println("Client committed a sin: " + err.Error())
					record.LogDisconnection(err.(ClientError).client)
				default:
					log.Println(err)
				}
			case <-abortChan:
				return
			case state := <-stateChan:
				// a state change has occurred
				record.LogState(state)

				if state.Finished() {
					record.LogResult(state)
					log.Printf("Result: %s\n", state.Result())
					log.Println("Game over.")
					return
				}
			}
		}
	}()

	return handler
}

// Listen for messages send from and received by this client in separate
// non-blocking goroutines.
func Listen(c GameClient) {
	go func() {
		for {
			broadcast := <-c.Send()
			err := websocket.JSON.Send(c.Conn(), &broadcast)
			if err != nil {
				c.Error() <- ClientError{err, c}
			}
		}
	}()

	go func() {
		for {
			var msg ClientMessage
			err := websocket.JSON.Receive(c.Conn(), &msg)
			if err != nil {
				c.Error() <- ClientError{err, c}
			}

			c.Receive() <- msg
		}
	}()
}

// A simple connection manager that forwards all connections along the
// connection channel in its handler.
type SimpleConnectionManager struct {
	Connections []*websocket.Conn
	ExitChans   map[*websocket.Conn]chan bool
}

func NewSimpleConnectionManager() *SimpleConnectionManager {
	return &SimpleConnectionManager{
		[]*websocket.Conn{},
		map[*websocket.Conn]chan bool{},
	}
}

func (m *SimpleConnectionManager) Handler(
	connChan chan *websocket.Conn,
	errChan chan error,
) websocket.Handler {

	return websocket.Handler(func(conn *websocket.Conn) {
		defer func() {
			err := conn.Close()
			if err != nil {
				errChan <- err
			}
		}()

		m.Connections = append(m.Connections, conn)
		connChan <- conn

		// keep the connection alive while the agents play the game
		exitChan := make(chan bool)
		m.ExitChans[conn] = exitChan
		<-exitChan

		conn.Close()
	})
}

func (m *SimpleConnectionManager) Close() {
	for _, c := range m.Connections {
		m.ExitChans[c] <- true
	}
}

// An authenticated client manager will require secret keys passed in for each
// client id. If a client does not pass in a valid key, or passes in a
// duplicate key, then it will be rejected.
type AuthenticatedClientManager struct {
	constructor   func(id string, conn *websocket.Conn) GameClient
	clients       []GameClient
	clientIds     []string
	clientSecrets []string
	timeout       time.Duration
}

// Create a new simple client manager. Give it a constructor to create clients
// from connections and also a list of client ids and secrets to expect.
func NewAuthenticatedClientManager(
	constructor func(id string, conn *websocket.Conn) GameClient,
	clientIds []string,
	clientSecrets []string,
	timeout time.Duration,
) *AuthenticatedClientManager {
	return &AuthenticatedClientManager{
		constructor,
		make([]GameClient, 0, len(clientIds)),
		clientIds,
		clientSecrets,
		timeout,
	}
}

func (m *AuthenticatedClientManager) Register(
	connChan chan *websocket.Conn,
) *sync.WaitGroup {

	var wg sync.WaitGroup
	wg.Add(1)

	watchdog := NewWatchdog(m.timeout)
	watchChan := watchdog.Watch()

	go func() {
		for {
			select {
			case conn := <-connChan:
				log.Println("Client received")

				id, err := m.Validate(conn)
				if err != nil {
					// Client sent invalid authorization parameters
					log.Println("Client rejected: " + err.Error())
					conn.Close()
				} else {
					log.Println("Client accepted")
					client := m.constructor(id, conn)
					m.clients = append(m.clients, client)
				}

				if len(m.clients) == cap(m.clients) {
					wg.Done()
					return
				}
			case <-watchChan:
				// Quit the goroutine if the watchdog times out.
				wg.Done()
				return
			}
		}
	}()

	return &wg
}

func (m *AuthenticatedClientManager) Validate(
	conn *websocket.Conn,
) (string, error) {
	secret := conn.Request().Header.Get("Authorization")
	if secret == "" {
		// no key sent by client
		return "", errors.New("Secret is required.")
	}

	// check if secret is in the list
	valid := false
	for i, k := range m.clientSecrets {
		if k == secret {
			valid = true
			id := m.clientIds[i]
			// remove used secrets
			m.clientSecrets = append(m.clientSecrets[:i], m.clientSecrets[i+1:]...)
			m.clientIds = append(m.clientIds[:i], m.clientIds[i+1:]...)

			return id, nil
		}
	}

	if valid == false {
		return "", errors.New("Invalid secret.")
	}

	return "", nil
}

func (m *AuthenticatedClientManager) Clients() []GameClient {
	return m.clients
}

func (m *AuthenticatedClientManager) Ready() bool {
	return len(m.clients) == cap(m.clients)
}

// The game recorder records the game state every time it changes and whether
// a client connects as expected or disconnects unexpectedly. This allows
// the sandbox service to adequately punish clients which are not well-behaved,
// and send game results to the scoreboard service.
type SimpleGameRecorder struct {
	StateLog      *os.File
	ResultLog     *os.File
	ConnectLog    *os.File
	DisconnectLog *os.File
}

func NewSimpleGameRecorder() (*SimpleGameRecorder, error) {
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := os.FileMode(0600)
	stateLog, err := os.OpenFile(StateLogFile, flags, perm)
	if err != nil {
		return nil, err
	}
	resultLog, err := os.OpenFile(ResultLogFile, flags, perm)
	if err != nil {
		return nil, err
	}
	connectLog, err := os.OpenFile(ConnectLogFile, flags, perm)
	if err != nil {
		return nil, err
	}
	disconnectLog, err := os.OpenFile(DisconnectLogFile, flags, perm)
	if err != nil {
		return nil, err
	}

	return &SimpleGameRecorder{
		stateLog,
		resultLog,
		connectLog,
		disconnectLog,
	}, nil
}

func (r *SimpleGameRecorder) LogState(s GameState) error {
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

func (r *SimpleGameRecorder) LogResult(s GameState) error {
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

func (r *SimpleGameRecorder) LogConnection(c GameClient) error {
	_, err := r.ConnectLog.WriteString(c.Id() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (r *SimpleGameRecorder) LogDisconnection(c GameClient) error {
	_, err := r.DisconnectLog.WriteString(c.Id() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (r *SimpleGameRecorder) Close() error {
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
	ch      chan bool
	timer   *time.Timer
}

func NewWatchdog(timeout time.Duration) *Watchdog {
	return &Watchdog{timeout, make(chan bool), nil}
}

// Start the watchdog on a separate goroutine. Will call Done() on the given
// waitgroup when it times out unless it is stopped before the timer is done.
func (w *Watchdog) Watch() chan bool {
	w.timer = time.AfterFunc(w.timeout, func() {
		w.ch <- true
	})
	return w.ch
}

// Stop the watchdog from sending an error when the timeout is reached.
func (w *Watchdog) Stop() {
	if w.timer != nil {
		w.timer.Stop()
	}
}
