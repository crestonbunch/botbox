package server

type GameState interface {
	// Return a JSON-serializable representatino of actions that the player p can
	// make in the current state.
	Actions(p int) interface{}

	// Commit the given action a (in JSON) for player p.
	Do(p int, a interface{})

	// Get a JSON-serializable state of the game, as seen by player p
	View(p int) interface{}

	// Check if the game is over yet or not.
	Finished() bool
}

type GameMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type TurnMessage struct {
	Turn    int         `json:"turn"`
	Player  int         `json:"player"`
	Actions interface{} `json:"actions"`
	State   interface{} `json:"state"`
}
