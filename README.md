Bot Box
=======
Bot Box is a customizable server meant for battling AIs against each other. It
comes in multiple parts:

* A sandboxed game server with scheduler that plays bots against each other by
  communicating via websockets.
* A framework for creating game rules and constraints to be played by bots in
real-time or alternating turns.
* SDKs for a number of languages to write bots.
* A web visualizer for watching games (games will need a view component) and
tracking high scores.

Progress
========

Currently the following is implemented and working:

* Example Tron game and websocket server with a Python client SDK.

We still need a script to sandbox games and a scheduler to schedule them and a
storage service to store results and a web service to watch games and a user
service to manage users and upload agents, etc.

So, yeah... lots of work left.

Usage
=====

Right now you can run ```go run main.go``` from games/tron/server.
This starts the Tron server and waits for clients.

Install the Tron SDK from games/tron/sdk/python using ```python setup.py develop```

Then write a simple Tron agent, e.g.:

```
import botbox_tron
import random

def move(p, actions, state):
    """A trivial agent that randomly picks an action."""

    # find moves that are safe to make
    safe = botbox_tron.safe_moves(p, state)
    # find the intersection of valid and safe moves
    actions = [a for a in actions if a in safe]

    if actions:
        # pick a valid, safe action
        return random.choice(actions)
    else:
        # can't do anything
        return

# Start the agent with our simple move function
botbox_tron.start(move)
```
and run two instances of it to watch them play each other!
