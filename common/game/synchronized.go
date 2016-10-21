package game

import (
	"errors"
	"golang.org/x/net/websocket"
	"log"
	"sync"
	"time"
)

type SynchronizedGameClient struct {
	id       string
	conn     *websocket.Conn
	watchdog *Watchdog
	send     chan ServerMessage
	receive  chan ClientMessage
	err      chan ClientError
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

func (c *SynchronizedGameClient) Send() chan ServerMessage {
	return c.send
}

func (c *SynchronizedGameClient) Receive() chan ClientMessage {
	return c.receive
}

func (c *SynchronizedGameClient) Error() chan ClientError {
	return c.err
}

type SynchronizedStateManager struct {
	state   GameState
	timeout time.Duration
}

func NewSynchronizedStateManager(
	game GameState, timeout time.Duration,
) *SynchronizedStateManager {
	return &SynchronizedStateManager{game, timeout}
}

func (m *SynchronizedStateManager) NewClient(
	id string, conn *websocket.Conn,
) GameClient {
	return &SynchronizedGameClient{
		id,
		conn,
		NewWatchdog(m.timeout),
		make(chan ServerMessage),
		make(chan ClientMessage),
		make(chan ClientError),
	}
}

// Synchronizes gameplay so both players make moves at the same time. If a
// player does not make a move in the allotted timeframe, then it its turn
// is skipped and a timeout error is sent along the error channel. Game states
// may punish a client by doing something if the action received is the empty
// string.
func (m *SynchronizedStateManager) Play(
	clients []GameClient,
	stateChan chan GameState,
	errChan chan error,
) *sync.WaitGroup {

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for !m.state.Finished() {
			// wait for actions from every player to commit them simultaneously
			actions := make([]string, len(clients))
			// block for all players and queue up their actions
			for i, c := range clients {
				watchCh := c.Watchdog().Watch()
				log.Println("Sending message to client " + c.Id())
				// note that if an error is returned, then the action will be the empty
				// string, so a state can kill a player if the empty string is received
				// to punish bad players
				select {
				case c.Send() <- ServerMessage{i, m.state.Actions(i), m.state.View(i)}:
				case <-watchCh:
					errChan <- errors.New("Client send timeout")
				}

				select {
				case msg := <-c.Receive():
					actions[i] = msg.Action
				case err := <-c.Error():
					errChan <- err
				case <-watchCh:
					errChan <- errors.New("Client receive timeout")
				}
				c.Watchdog().Stop()
				log.Println("Got action '" + actions[i] + "' from client " + c.Id())
			}
			// commit actions simultaneously
			for i, a := range actions {
				m.state.Do(i, a)
			}
			stateChan <- m.state
			log.Println("Committed actions.")
		}

		wg.Done()
	}()

	return &wg
}
