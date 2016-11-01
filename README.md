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
* Sandbox service which spawns isolated Docker containers and plays games
  started via an HTTP API.
* Database container with empty PostgreSQL database

Remaining work:

* Finish the web API
* Build a web interface
* Create a scheduler service for matchmaking
* Continuous deployment and upgrade scripts

Usage
=====

Right now from ```games/tron/server``` you can run

 ```go run main.go --ids "1 2" --secrets "s1 s2"```

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

Deploying
=========

Right now there is a simple ./deploy.sh script you can run to start up the
docker containers.
