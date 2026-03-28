package engine

import "testing"

func TestEvaluateMatchesRecomputedStateAcrossMoves(t *testing.T) {
	const fen = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

	var pos PositionNG
	pos.Set(fen)

	if got, want := pos.Evaluate(), pos.evaluateNoCache(); got != want {
		t.Fatalf("initial evaluate mismatch: got %d want %d", got, want)
	}

	var states [16]StateInfo
	var played [16]MoveNG
	plies := 0

	for ; plies < len(states); plies++ {
		var moves [MAX_MOVES]MoveNG
		size := pos.GenerateLEGAL(moves[:])
		if size == 0 {
			break
		}
		mv := moves[(plies*7+3)%int(size)]
		played[plies] = mv
		pos.DoMove(mv, &states[plies])

		if !pos.PosIsOk() {
			t.Fatalf("position invalid after ply %d", plies+1)
		}
		if got, want := pos.Evaluate(), pos.evaluateNoCache(); got != want {
			t.Fatalf("after ply %d evaluate mismatch: got %d want %d", plies+1, got, want)
		}
	}

	for i := plies - 1; i >= 0; i-- {
		pos.UndoMove(played[i])
		if !pos.PosIsOk() {
			t.Fatalf("position invalid after undo %d", i)
		}
		if got, want := pos.Evaluate(), pos.evaluateNoCache(); got != want {
			t.Fatalf("after undo %d evaluate mismatch: got %d want %d", i, got, want)
		}
	}
}

func TestEvaluateRewardsAdvancedPawnPressure(t *testing.T) {
	var advanced PositionNG
	advanced.Set("4k4/9/9/4P4/9/9/9/9/9/5K3 w - - 0 1")

	var passive PositionNG
	passive.Set("4k4/9/9/9/9/9/4P4/9/9/5K3 w - - 0 1")

	if advanced.Evaluate() <= passive.Evaluate() {
		t.Fatalf("expected advanced pawn to evaluate higher: advanced=%d passive=%d", advanced.Evaluate(), passive.Evaluate())
	}
}
