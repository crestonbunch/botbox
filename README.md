Bot Box
=======
Bot Box is a customizable server meant for battling AIs against each other. It
comes in multiple parts:

* A sandboxed game server with scheduler that plays bots against each other by
  communicating via TCP over UNIX sockets.
* A framework for creating game rules and constraints to be played by bots in
real-time or alternating turns.
* API wrappers for a number of languages to write bots.
* A web visualizer for watching games (games will need a view component) and
tracking high scores.
