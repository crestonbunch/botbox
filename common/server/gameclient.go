package server

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
)

// This is a client for a synchronized game server. Clients must decided on an
// action to take each turn.
// When a new turn comes around, the server will send a message that looks like:
// {turn: 10, player: 1, actions: [..], state: {..}}
// Clients must send a response that looks like:
// looks like: {type: "do", payload: "..."}, where the payload is the
// specific action to take.
type SynchronizedGameClient struct {
	conn   *websocket.Conn
	server *SynchronizedGameServer
	index  int
	action chan interface{}
	turn   chan *TurnMessage
	done   chan bool
}

func NewSynchronizedGameClient(ws *websocket.Conn, server *SynchronizedGameServer) *SynchronizedGameClient {
	return &SynchronizedGameClient{
		ws,
		server,
		0,
		make(chan interface{}, 1),
		make(chan *TurnMessage),
		make(chan bool),
	}
}

func (c *SynchronizedGameClient) SignalTurn(num int, player int, actions interface{}, state interface{}) {
	c.turn <- &TurnMessage{num, player, actions, state}
}

func (c *SynchronizedGameClient) Listen() {
	go func() {
		for {
			select {
			case msg := <-c.turn:
				err := websocket.JSON.Send(c.conn, msg)
				if err != nil {
					fmt.Println("Error:", err)
				}
				//case done := <-c.done:
				//return
			}
		}
	}()

	for {
		select {
		//case done := <-c.done:
		//return
		default:
			// receive a message from the client
			var msg GameMessage
			err := websocket.JSON.Receive(c.conn, &msg)

			if err == io.EOF {
				c.done <- true
			} else if err != nil {
				fmt.Println("Error:", err)
				c.server.err <- err
			} else {
				switch string(msg.Type) {
				case "do":
					// queue an action for this turn
					c.action <- msg.Payload
				}
			}
		}
	}
}
