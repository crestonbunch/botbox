package game

func mockTwoPlayerGame() *mockState {
	return &mockState{[]int{0, 0}}
}

type mockGameRecorder struct{}

func (r *mockGameRecorder) LogState(s GameState) error {
	return nil
}

func (r *mockGameRecorder) LogResult(s GameState) error {
	return nil
}

func (r *mockGameRecorder) LogConnection(c GameClient) error {
	return nil
}

func (r *mockGameRecorder) LogDisconnection(c GameClient) error {
	return nil
}

func (r *mockGameRecorder) Close() error {
	return nil
}

type mockState struct {
	Players []int `json:"players"`
}

func (s *mockState) Actions(p int) interface{} {
	return []string{"1", "2", "3"}
}

func (s *mockState) Do(p int, a string) {
	if a == "1" {
		s.Players[p] += 1
	} else if a == "2" {
		s.Players[p] += 2
	} else if a == "3" {
		s.Players[p] += 3
	}
}

func (s *mockState) View(p int) interface{} {
	return s
}

func (s *mockState) Finished() bool {
	for _, p := range s.Players {
		if p >= 10 {
			return true
		}
	}

	return false
}

func (s *mockState) Result() []int {
	result := []int{}
	max := -1
	for _, p := range s.Players {
		if p > max {
			max = p
		}
	}
	count := 0
	for _, p := range s.Players {
		if p == max {
			result = append(result, ResultWin)
			count += 1
		} else {
			result = append(result, ResultLoss)
		}
	}
	if count > 1 {
		for i, p := range s.Players {
			if p == max {
				result[i] = ResultTie
			}
		}
	}

	return result
}
