package tron

import (
	"testing"
)

func TestValidMoves(t *testing.T) {
	state := NewTwoPlayerTron(32, 32)

	if state.Players[0].X != 0 && state.Players[0].Y != 0 {
		t.Error("Player 1 is not at (0,0)")
		return
	}
	if state.Players[1].X != 31 && state.Players[1].Y != 31 {
		t.Error("Player 2 is not at (31,31)")
		return
	}

	p1Actions := state.Actions(0).([]string)
	if len(p1Actions) != 2 && p1Actions[0] != DirectionSouth && p1Actions[1] != DirectionEast {
		t.Error("Player 1 actions are not ['south', 'east']")
		return
	}
	p2Actions := state.Actions(1).([]string)
	if len(p2Actions) != 2 && p2Actions[0] != DirectionNorth && p1Actions[1] != DirectionWest {
		t.Error("Player 2 actions are not ['north', 'west']")
		return
	}

	state.Do(0, "south")
	state.Do(1, "west")

	if state.Players[0].X != 0 && state.Players[0].Y != 1 {
		t.Error("Player 1 is not at (0,1)")
		return
	}
	if state.Players[1].X != 30 && state.Players[1].Y != 31 {
		t.Error("Player 2 is not at (30,31)")
		return
	}

}
