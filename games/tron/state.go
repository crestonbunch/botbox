// This file contains state structures for a Tron game. These tron games are
// a 2-dimensional grid with arbitrary numbers of players.
package tron

import "strconv"

const (
	DirectionNorth = "north"
	DirectionEast  = "east"
	DirectionSouth = "south"
	DirectionWest  = "west"
)

// The state of the tron world, holds lists of coordinates for an arbitrary
// number of players. The top left cell is coordinate (0,0) like screen coords.
type TronState struct {
	// A map of x, y coordinates to player indices, indicating which cells are
	// part of which player's "trail". The keys are strings due to those being
	// the only supported type by the JSON library.
	Cells map[string]map[string]int `json:"cells"`
	// The players slice is simply a lookup of a player's position, where each
	// player is tracked by its index in this list. If the player's position is
	// (-1, -1) then the player is dead.
	Players []TronCoord `json:"players"`
	// track the current direction the players are traveling in
	Directions []string
	// The width and height of the game grid.
	Width  int `json:"w"`
	Height int `json:"h"`
}

type TronCoord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Build a blank tron world with 2 players.
func NewTwoPlayerTron(w, h int) *TronState {
	// the map starts empty
	cells := map[string]map[string]int{}
	// players start in opposite corners
	players := []TronCoord{TronCoord{0, 0}, TronCoord{w - 1, h - 1}}
	playersDir := []string{DirectionSouth, DirectionNorth}
	return &TronState{cells, players, playersDir, w, h}
}

// Returns the possible actions an agent can make given the current state. Each
// action is a unit tuple of directions that the player can travel in.
func (s *TronState) Actions(p int) interface{} {
	x, y := s.Players[p].X, s.Players[p].Y

	a := []string{}
	if x < 0 && y < 0 {
		// player is dead
		return a
	}

	if y > 0 && s.Directions[p] != DirectionSouth {
		// player can turn north
		a = append(a, DirectionNorth)
	}

	if y < s.Height-1 && s.Directions[p] != DirectionNorth {
		// player can turn south
		a = append(a, DirectionSouth)
	}

	if x > 0 && s.Directions[p] != DirectionEast {
		// player can turn west
		a = append(a, DirectionWest)
	}

	if x < s.Width-1 && s.Directions[p] != DirectionWest {
		// player can turn east
		a = append(a, DirectionEast)
	}

	return a
}

// Commit an action for a player. The direction will be sent as a generic
// interface, and casted into the expected string
func (s *TronState) Do(p int, dir interface{}) {

	// validate type
	switch dir.(type) {
	case nil:
		// player has not made a move, the punishment is death
		s.Kill(p)
		return
	}

	a := dir.(string)

	// to the integer to string conversion for coordinates
	x := strconv.Itoa(s.Players[p].X)
	y := strconv.Itoa(s.Players[p].Y)
	// verify this action is a valid move
	if s.Validate(p, a) {
		// change player direction
		s.Directions[p] = a
		// grow the player's trail
		if v, ok := s.Cells[x]; ok {
			v[y] = p
		} else {
			// Y map doesn't exist yet, create it
			s.Cells[strconv.Itoa(s.Players[p].X)] = map[string]int{y: p}
		}
		// move player
		switch a {
		case DirectionNorth:
			s.Players[p].Y -= 1
		case DirectionEast:
			s.Players[p].X += 1
		case DirectionSouth:
			s.Players[p].Y += 1
		case DirectionWest:
			s.Players[p].X -= 1
		}

		// update x and y values
		x = strconv.Itoa(s.Players[p].X)
		y = strconv.Itoa(s.Players[p].Y)
		if v, ok := s.Cells[x]; ok {
			if _, ok := v[y]; ok {
				// if the player ran over a tail -- kill him
				s.Kill(p)
			}
		}

		for i, o := range s.Players {
			if i != p && o.X == s.Players[p].X && o.Y == s.Players[p].Y {
				// two or more players overlap, kill both of them
				// TODO: this does not work if 3 or more players overlap
				o.X = -1
				o.X = -1
				s.Kill(p)
			}
		}
	} else {
		// player has made an invalid action -- the punishment is death
		s.Kill(p)
	}
}

// Tron is a perfect-information game, so return the state regardless of player.
func (s *TronState) View(p int) interface{} {
	return s
}

// Check if the game is over. This happens if all but one player is dead.
// TODO: Currently this assumes only a 2 player game
func (s *TronState) Finished() bool {
	for _, p := range s.Players {
		if p.X == -1 && p.Y == -1 {
			return true
		}
	}
	return false
}

// Check if an action is valid.
func (s *TronState) Validate(p int, a string) bool {
	for _, b := range s.Actions(p).([]string) {
		if a == b {
			return true
		}
	}
	return false
}

// Kill a player.
func (s *TronState) Kill(p int) {
	s.Players[p].X = -1
	s.Players[p].Y = -1
}
