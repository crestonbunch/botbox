import websocket
import json
import thread

WS_SERVER_URL = 'ws://localhost:12345/'

def start(turn_handler):
    """Start the client listening to the game. Pass in a function
    that accepts the available actions and the current state of the game,
    and returns the action to take. The SDK will handle the rest."""
    ws = websocket.WebSocketApp(
        WS_SERVER_URL,
        on_open = _on_open(ws),
        on_message = lambda ws, msg: _on_message(ws, msg, turn_handler),
        on_error = _on_error(ws, msg),
        on_close = _on_close(ws)
    )
    websocket.create_connection(SERVER_URL)

    ws.run_forever()

def _on_message(ws, msg, turn_handler):
    """This is a private method that handles incoming messages from
    the websocket, passes the turn information to an agent's turn
    handler, and then passes the result back to the server."""

    def x():
        parsed = json.dumps(msg)
        actions = parsed['actions']
        state = parsed['state']

        action = turn_handler(actions, state)

        ws.send(json.loads(action))

    thread.start_new_thread(x, ())

def _on_open(ws):
    print('Connection opened')

def _on_error(ws, msg):
    print('Error:', msg)

def _on_close(ws):
    print('Connection closed.')
