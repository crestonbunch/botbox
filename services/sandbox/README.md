Sandbox Service
===============

The sandbox service is responsible for spinning up Docker containers to manage a
single match. The containers it creates are:

* A container for the game server which manages the game (e.g., Tron server)
* A separate container for each agent.

The server will spin up a local server, and all the agents will be configured to
connect to it, and the game will kick-off.

Results curated by the sandbox service will be sent to the scoreboard service.

Running Scripts
===============

The sandbox looks for a 'run.sh' script in the source directory for each
client and server. A default is provided that will look for typical run files
in each language, however this script can be overwritten by a custom one
inside the source directory provided by the user. So in case the default
run.sh does not meet your needs, you can provide a custom one.

Usage
-----

```$ cd server/```

```$ ./build.sh && ./run.sh```

This will start an HTTP server listening on port 8080.
