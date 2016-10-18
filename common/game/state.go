package game

const (
	ResultWin  = 1
	ResultTie  = 0
	ResultLoss = -1
)

type GameState interface {
	// Return a JSON-serializable representation of actions that the player p can
	// make in the current state.
	Actions(p int) interface{}

	// Commit the given action a (in JSON) for player p.
	Do(p int, a string)

	// Get a JSON-serializable state of the game, as seen by player p
	View(p int) interface{}

	// Check if the game is over yet or not.
	Finished() bool

	// Decide the result of the match. Each player gets a score +1, 0, -1
	Result() []int
}
