import websocket
import json
import _thread
import sys

WS_SERVER_URL = 'ws://localhost:12345/'

def start(turn_handler):
    """Start the client listening to the game. Pass in a function
    that accepts the available actions and the current state of the game,
    and returns the action to take. The SDK will handle the rest.
    Checks if any command-line arguments are passed when running,
    if there are any, they are assumed to be client keys that are
    sent to the server for connecting."""


    headers = {'Authentication': sys.argv[1]} if len(sys.argv) > 1 else []
    if headers: print("Using headers", headers)
    ws = websocket.WebSocketApp(
        WS_SERVER_URL,
        on_open = _on_open,
        on_message = lambda ws, msg: _on_message(ws, msg, turn_handler),
        on_error = _on_error,
        on_close = _on_close,
        header = headers
    )

    ws.run_forever()

def _on_message(ws, msg, turn_handler):
    """This is a private method that handles incoming messages from
    the websocket, passes the turn information to an agent's turn
    handler, and then passes the result back to the server."""

    def x():
        parsed = json.loads(msg)
        actions = parsed['actions']
        state = parsed['state']

        for y in range(state['h']):
            for x in range(state['w']):
                x_str, y_str = str(x), str(y)
                if x_str in state['cells'] and y_str in state['cells'][x_str]:
                    sys.stdout.write(str(state['cells'][x_str][y_str]))
                elif state['players'][0]['x'] == x and state['players'][0]['y'] == y:
                    sys.stdout.write('A')
                elif state['players'][1]['x'] == x and state['players'][1]['y'] == y:
                    sys.stdout.write('B')
                else:
                    sys.stdout.write(' ')
            sys.stdout.write('\n')

        action = turn_handler(actions, state)
        response = {"type":"do", "payload":action}

        ws.send(json.dumps(response))
        print("Sent", json.dumps(response))

    _thread.start_new_thread(x, ())

def _on_open(ws):
    print('Connection opened')

def _on_error(ws, msg):
    print('Error:', msg)

def _on_close(ws):
    print('Connection closed.')
