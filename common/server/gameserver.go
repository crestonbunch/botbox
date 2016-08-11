package server

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
)

// A generic websocket server for games that are turn-based. Players make
// turns in the order that they connect in. Moves are made alternating, so
// that each turn only a single player moves.
// TODO: implement
type TurnBasedGameServer struct {
}

// A generic websocket server for games that are played in real time. Players
// must request the current state / available actions at their discretion, and
// can signal an action to the server at any time. Player actions are queued
// and validated in the order received, so players should never be able to make
// invalid moves if the state has changed since they committed an action.
// TODO: implement
type RealTimeGameServer struct {
}

// A generic websocket server for games to use (if they want.)
// They must provide a state struct that implements the
// GameState interface. Synchronized means that the players move simultaneously,
// on the same turn, and have no idea what move the other player will make.
// TODO: pass a secret key for each client that is to be in the game so that
// a malicious client does not try to spawn additional connections to the server
type SynchronizedGameServer struct {
	state   GameState
	clients []*SynchronizedGameClient
	turn    int
	delete  chan *SynchronizedGameClient
	done    chan bool
	err     chan error
}

func NewSynchronizedGameServer(state GameState, players int) *SynchronizedGameServer {
	return &SynchronizedGameServer{
		state,
		make([]*SynchronizedGameClient, 0, players),
		0,
		make(chan *SynchronizedGameClient),
		make(chan bool),
		make(chan error),
	}
}

func (s *SynchronizedGameServer) Register(c *SynchronizedGameClient) {
	c.index = len(s.clients)
	s.clients = append(s.clients, c)
}

func (s *SynchronizedGameServer) Start(path string) {
	// handler to register connected clients
	http.Handle(path, websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				s.err <- err
			}
		}()
		fmt.Println("Client connected")
		client := NewSynchronizedGameClient(ws, s)
		s.Register(client)
		client.Listen()
	}))

	fmt.Println("Waiting for clients to connect")

	for len(s.clients) < cap(s.clients) {
		// block until all clients have connected
	}

	fmt.Println("All clients connected")

	s.Listen()
}

func (s *SynchronizedGameServer) Stop() {
	fmt.Println("Sending stop signal.")
	s.done <- true
}

func (s *SynchronizedGameServer) Listen() {

	// concurrently listen for player actions
	go func() {
		for {
			// check if the game is over
			if s.state.Finished() {
				s.Stop()
			} else {
				// broadcast the current turn
				for i, c := range s.clients {
					c.SignalTurn(s.turn, s.state.Actions(i), s.state.View(i))
				}
			}

			select {
			case <-s.done:
				// broadcast the end results
				for i, c := range s.clients {
					c.SignalTurn(s.turn, []interface{}{}, s.state.View(i))
				}
				return
			default:
				// wait for actions from every player to commit them simultaneously
				actions := make([]interface{}, len(s.clients))
				// block for all players and queue up their actions
				for i, c := range s.clients {
					actions[i] = <-c.action
				}
				// commit actions simultaneously
				for i, a := range actions {
					fmt.Println("Commiting action", a, "for player", i)
					s.state.Do(i, a)
				}
			}
		}
	}()

	// handle other requests from clients
	for {
		select {
		case <-s.done:
			return
			//case client := <-s.delete:
			// not implemented
			//case err := <-s.err:
			// not implemented
		}
	}
}
