// This file contains state structures for a Tron game. These tron games are
// a 2-dimensional grid with arbitrary numbers of players.
package tron

// The state of the tron world, holds lists of coordinates for an arbitrary
// number of players.
type TronState struct {
	// Each cell that is non-empty will be in this map. Each cell will have
	// a number corresponding to the player that occupied that cell.
	cells map[TronCoord]int `json:"cells"`
	// The players slice is simply a lookup of a player's position, where each
	// player is tracked by its index in this list. If the player's position is
	// (-1, -1) then the player is dead.
	players []TronCoord `json:"players"`
	// The width and height of the game grid.
	w int `json:"w"`
	h int `json:"h"`
}

type TronCoord struct {
	x int `json: "x"`
	y int `json: "y"`
}

// Build a blank tron world with 2 players.
func NewTwoPlayerTron(w, h int) *TronState {
	return &TronState{map[TronCoord]int{}, []TronCoord{}, w, h}
}

// Returns the possible actions an agent can make given the current state. Each
// action is a unit tuple of directions that the player can travel in.
func (s *TronState) Actions(p int) interface{} {
	x, y := s.players[p].x, s.players[p].y

	a := []TronCoord{}
	if x < 0 && y < 0 {
		// player is dead
		return a
	}
	if x < s.w-1 {
		// player can move right
		a = append(a, TronCoord{1, 0})
	}
	if x > 0 {
		// player can move left
		a = append(a, TronCoord{-1, 0})
	}
	if y < s.h-1 {
		// player can move down
		a = append(a, TronCoord{0, 1})
	}
	if y > 0 {
		// player can move up
		a = append(a, TronCoord{0, -1})
	}

	return a
}

// Commit an action for a player. Returns a boolean whether or not the
// action was committed. The coord will be a map parsed from JSON since it gets
// passed around as a generic interface (but was originally a TronCoord type)
func (s *TronState) Do(p int, coord interface{}) {
	a = TronCoord{a["x"], a["y"]}
	// verify this action is a valid move
	if s.Validate(p, a) {
		// grow the player's trail
		s.cells[s.players[p]] = p
		// move the player
		s.players[p].x += a.x
		s.players[p].y += a.y

		if _, ok := s.cells[s.players[p]]; ok {
			// if the player ran over a tail -- kill him
			s.players[p].x = -1
			s.players[p].y = -1
		}

		for i, o := range s.players {
			if i != p && o.x == s.players[p].x && o.y == s.players[p].y {
				// two or more players overlap, kill both of them
				// TODO: this does not work if 3 or more players overlap
				o.x = -1
				o.y = -1
				s.players[p].x = -1
				s.players[p].y = -1
			}
		}

		return true
	}
	return false
}

// Check if the game is over. This happens if all but one player is dead.
// TODO: Currently this assumes only a 2 player game
func (s *TronState) Finished() bool {
	for _, p := range s.players {
		if p.x == -1 && p.y == -1 {
			return true
		}
	}
	return false
}

// Check if an action is valid.
func (s *TronState) Validate(p int, a TronCoord) bool {
	for _, b := range s.Actions(p) {
		if a.x == b.x && a.y == b.y {
			return true
		}
	}
	return false
}
