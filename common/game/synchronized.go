package game

import (
	"errors"
	"golang.org/x/net/websocket"
	"log"
)

func NewSynchronizedGameServer(
	game GameState,
	clients []string,
	keys []string,
) (*GameServer, error) {
	writer, err := NewGameRecorder(
		StateLogFile,
		ResultLogFile,
		ConnectLogFile,
		DisconnectLogFile,
	)
	if err != nil {
		return nil, err
	}
	return NewGameServer(
		NewSimpleConnectionManager(len(keys)),
		NewAuthenticatedClientManager(NewSynchronizedGameClient, clients, keys),
		NewSynchronizedStateManager(game),
		writer,
	), nil
}

type SynchronizedGameClient struct {
	id       string
	conn     *websocket.Conn
	watchdog *Watchdog
}

func (c *SynchronizedGameClient) Id() string {
	return c.id
}

func (c *SynchronizedGameClient) Conn() *websocket.Conn {
	return c.conn
}

func (c *SynchronizedGameClient) Watchdog() *Watchdog {
	return c.watchdog
}

func NewSynchronizedGameClient(id string, conn *websocket.Conn) GameClient {
	return &SynchronizedGameClient{id, conn, NewWatchdog(MoveTimeout)}
}

type SynchronizedStateManager struct {
	state GameState
}

func NewSynchronizedStateManager(game GameState) *SynchronizedStateManager {
	return &SynchronizedStateManager{game}
}

func (m *SynchronizedStateManager) Play(
	clients []GameClient,
	stateChan chan GameState,
	errChan chan error,
) {

	for !m.state.Finished() {
		// broadcast the current turn
		for i, c := range clients {
			msg := ServerMessage{i, m.state.Actions(i), m.state.View(i)}
			err := websocket.JSON.Send(c.Conn(), msg)
			if err != nil {
				// There was a problem communicating with a specific client
				errChan <- ClientError{err, c}
			}
		}
		log.Println("Broadcast turn.")

		// wait for actions from every player to commit them simultaneously
		actions := make([]string, len(clients))
		// block for all players and queue up their actions
		for i, c := range clients {
			action, err := m.Move(i, c)
			// note that if an error is returned, then the action will be the empty
			// string, so a state can kill a player if the empty string is received
			// to punish bad players
			actions[i] = action
			if err != nil {
				errChan <- ClientError{err, c}
			}
			log.Println("Recevied action '" + action + "' from client " + c.Id())
		}
		// commit actions simultaneously
		for i, a := range actions {
			m.state.Do(i, a)
		}
		stateChan <- m.state
		log.Println("Committed actions.")
	}
}

// Make a move for a client, or return "" and an error if the client did not
// respond.
func (m *SynchronizedStateManager) Move(i int, c GameClient) (string, error) {
	errChan := make(chan error)
	msgChan := make(chan ClientMessage)
	// The client took too long to make a move.
	c.Watchdog().Start(
		errors.New("Client move timeout."),
		errChan,
	)

	go func() {
		var msg ClientMessage
		err := websocket.JSON.Receive(c.Conn(), &msg)
		if err != nil {
			// There was a problem communicating with a specific client
			errChan <- err
		}

		msgChan <- msg
	}()

	// Block until the player makes a move, or an error occurs.
	select {
	case err := <-errChan:
		c.Watchdog().Stop()
		return "", err
	case msg := <-msgChan:
		c.Watchdog().Stop()
		return msg.Action, nil
	}
}
