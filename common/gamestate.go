package botbox

type GameState interface {
	// Return a JSON-serializable representatino of actions that the player p can
	// make in the current state.
	Actions(p int) interface{}

	// Commit the given action a (in JSON) for player p.
	Do(p int, a interface{})

	// Get the visible state in JSON format for a given player.
	View(p int) []byte

	// Check if the game is over yet or not.
	Finished() bool
}

type GameMessage struct {
	Type    []byte      `json:"type"`
	Payload interface{} `json:"payload"`
}

type TurnMessage struct {
	Turn    int         `json:"turn"`
	Actions interface{} `json:"actions"`
	State   interface{} `json:"state"`
}
