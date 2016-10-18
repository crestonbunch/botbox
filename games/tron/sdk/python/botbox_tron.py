import websocket
import json
import _thread
import sys
import os

WS_SERVER_SCHEME = 'ws'
WS_SERVER_URL = 'localhost'
WS_SERVER_PORT = '12345'

def safe_moves(p, state):
    """Determine what moves are safe for a player to make. Returns a list of
    valid actions that player p can make in the given state."""

    x, y = state['players'][p]['x'], state['players'][p]['y']

    moves = []
    actions = [(1, 0, 'east'),
            (-1, 0, 'west'),
            (0, -1, 'north'),
            (0, 1, 'south')]
    for dx, dy, move in actions:
        tx, ty = str(x + dx), str(y + dy)
        if tx not in state['cells'] or ty not in state['cells'][tx]:
            moves.append(move)

    return moves

def start(turn_handler):
    """Start the client listening to the game. Pass in a function
    that accepts the available actions and the current state of the game,
    and returns the action to take. The SDK will handle the rest.
    Checks if any command-line arguments are passed when running,
    if there are any, they are assumed to be client keys that are
    sent to the server for connecting."""

    if os.environ.get('BOTBOX_SECRET'):
        print('Using env secret:', os.environ['BOTBOX_SECRET'])
        headers = {'Authorization': os.environ['BOTBOX_SECRET']}
    elif len(sys.argv) > 1:
        print('Using cli secret:', sys.argv[1])
        headers = {'Authorization': sys.argv[1]}
    else:
        print('Using no authentication')
        headers = []

    # get the URL for the server from an environment variable if it is set,
    # otherwise use the default localhost
    if os.environ.get('BOTBOX_SERVER'):
        url = (WS_SERVER_SCHEME + '://'
            + os.environ['BOTBOX_SERVER'] + ':' + WS_SERVER_PORT)
    else:
        url = WS_SERVER_SCHEME + '://' + WS_SERVER_URL + ':' + WS_SERVER_PORT

    print("Connecting to:", url)

    ws = websocket.WebSocketApp(
        url,
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
        player = parsed['player']
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

        action = turn_handler(player, actions, state)
        response = {"action":action}

        ws.send(json.dumps(response))
        print("Sent", json.dumps(response))

    _thread.start_new_thread(x, ())

def _on_open(ws):
    print('Connection opened')

def _on_error(ws, msg):
    print('Error:', msg)

def _on_close(ws):
    print('Connection closed.')
