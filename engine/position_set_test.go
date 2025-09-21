package engine

import "testing"

const initialFen = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

func TestPositionSetResetsState(t *testing.T) {
	var pos PositionNG
	pos.Set(initialFen)

	moves := []string{"h2e2", "h9g7", "h0g2"}
	var states [3]StateInfo
	for i, moveStr := range moves {
		move, err := ParseUCIMove(&pos, moveStr)
		if err != nil {
			t.Fatalf("parse move %s: %v", moveStr, err)
		}
		pos.DoMove(move, &states[i])
	}

	if !pos.PosIsOk() {
		t.Fatal("position invalid after applying moves")
	}

	pos.Set(initialFen)

	if !pos.PosIsOk() {
		t.Fatal("position invalid after reset")
	}
	if pos.SideToMove != WHITE {
		t.Fatalf("expected white to move after reset, got %d", pos.SideToMove)
	}

	if _, err := ParseUCIMove(&pos, moves[0]); err != nil {
		t.Fatalf("move should be legal after reset: %v", err)
	}
}
