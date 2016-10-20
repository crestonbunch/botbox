package tron

import (
	"github.com/crestonbunch/botbox/common/game"
	"testing"
)

func TestValidMoves(t *testing.T) {
	state := NewTwoPlayerTron(32, 32)

	if state.Players[0].X != 0 && state.Players[0].Y != 0 {
		t.Error("Player 1 is not at (0,0)")
	}
	if state.Players[1].X != 31 && state.Players[1].Y != 31 {
		t.Error("Player 2 is not at (31,31)")
	}

	p1Actions := state.Actions(0).([]string)
	if len(p1Actions) != 2 && p1Actions[0] != DirectionSouth && p1Actions[1] != DirectionEast {
		t.Error("Player 1 actions are not ['south', 'east']")
	}
	p2Actions := state.Actions(1).([]string)
	if len(p2Actions) != 2 && p2Actions[0] != DirectionNorth && p1Actions[1] != DirectionWest {
		t.Error("Player 2 actions are not ['north', 'west']")
	}

	state.Do(0, "south")
	state.Do(1, "west")

	if state.Players[0].X != 0 && state.Players[0].Y != 1 {
		t.Error("Player 1 is not at (0,1)")
	}
	if state.Players[1].X != 30 && state.Players[1].Y != 31 {
		t.Error("Player 2 is not at (30,31)")
	}

}

func TestPlayer1Wins(t *testing.T) {
	state := NewTwoPlayerTron(32, 32)
	state.Do(0, "south")
	state.Do(1, "south")
	if !state.Finished() {
		t.Error("Game is not over!")
	}

	result := state.Result()
	if result[0] != game.ResultWin {
		t.Error("Player 1 did not win!")
	}

	if result[1] != game.ResultLoss {
		t.Error("Player 2 did not lose!")
	}
}

func TestPlayer2Wins(t *testing.T) {
	state := NewTwoPlayerTron(32, 32)
	state.Do(0, "north")
	state.Do(1, "north")
	if !state.Finished() {
		t.Error("Game is not over!")
	}

	result := state.Result()
	if result[0] != game.ResultLoss {
		t.Error("Player 1 did not lose!")
	}

	if result[1] != game.ResultWin {
		t.Error("Player 2 did not win!")
	}
}

func TestPlayer2Crashes(t *testing.T) {
	state := NewTwoPlayerTron(5, 5)
	state.Do(0, "south")
	state.Do(1, "north")
	state.Do(0, "east")
	state.Do(1, "north")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "east")
	state.Do(1, "north")
	if !state.Finished() {
		t.Error("Game is not over!")
	}

	result := state.Result()
	if result[0] != game.ResultWin {
		t.Error("Player 1 did not win!")
	}

	if result[1] != game.ResultLoss {
		t.Error("Player 2 did not lose!")
	}
}

func TestPlayer1Crashes(t *testing.T) {
	state := NewTwoPlayerTron(5, 5)
	state.Do(0, "south")
	state.Do(1, "north")
	state.Do(0, "east")
	state.Do(1, "north")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "south")
	state.Do(1, "west")
	if !state.Finished() {
		t.Error("Game is not over!")
	}

	result := state.Result()
	if result[0] != game.ResultLoss {
		t.Error("Player 1 did not lose!")
	}

	if result[1] != game.ResultWin {
		t.Error("Player 2 did not win!")
	}
}

func TestPlayersTie(t *testing.T) {
	state := NewTwoPlayerTron(5, 5)
	state.Do(0, "south")
	state.Do(1, "north")
	state.Do(0, "south")
	state.Do(1, "north")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "east")
	state.Do(1, "west")
	state.Do(0, "east")
	state.Do(1, "west")
	if !state.Finished() {
		t.Error("Game is not over!")
	}

	result := state.Result()
	if result[0] != game.ResultTie {
		t.Error("Player 1 did not tie!")
	}

	if result[1] != game.ResultTie {
		t.Error("Player 2 did not tie!")
	}
}
