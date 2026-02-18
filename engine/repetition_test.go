package engine

import "testing"

// helper to apply a sequence of UCI move strings to a position.
func applyMoves(t *testing.T, pos *PositionNG, moves []string, states []StateInfo) {
	t.Helper()
	for i, ms := range moves {
		m, err := ParseUCIMove(pos, ms)
		if err != nil {
			t.Fatalf("parse move %s (index %d): %v", ms, i, err)
		}
		pos.DoMove(m, &states[i])
	}
}

// TestRepetitionPerpetualCheckByUs — white perpetually checks, white to move
// after cycle → REP_LOSE.
func TestRepetitionPerpetualCheckByUs(t *testing.T) {
	// White: K f0, R a1.  Black: K e9.
	// Kings on f0 and e9 — different files, no facing.
	// R a1 does not attack e9 initially.
	fen := "4k4/9/9/9/9/9/9/9/R8/5K3 w - - 0 1"
	var pos PositionNG
	pos.Set(fen)

	states := make([]StateInfo, 12)
	moves := []string{
		"a1e1", // R a1-e1+ (check: e1 attacks e9 via e-file)
		"e9d9", // K e9-d9 (evasion; d9 doesn't face f0)
		"e1d1", // R e1-d1+ (check: d1 attacks d9 via d-file)
		"d9e9", // K d9-e9
		"d1e1", // R d1-e1+ (check) — cycle starts
		"e9d9", // K e9-d9
		"e1d1", // R e1-d1+ (check)
		"d9e9", // K d9-e9 — cycle complete
	}
	applyMoves(t, &pos, moves, states)

	if !pos.IsRepetition() {
		t.Fatal("expected repetition")
	}

	// WHITE to move. White checked on all its moves → REP_LOSE.
	result := pos.ClassifyRepetition()
	if result != REP_LOSE {
		t.Fatalf("expected REP_LOSE, got %d", result)
	}
}

// TestRepetitionPerpetualCheckByBlack — black perpetually checks, black to
// move after cycle → REP_LOSE for black.
func TestRepetitionPerpetualCheckByBlack(t *testing.T) {
	// Black: K f9, r a1.  White: K e0.
	// Kings on f9 and e0 — different files, no facing.
	// r a1 does not attack e0 initially (different file/rank).
	fen := "5k3/9/9/9/9/9/9/9/r8/4K4 b - - 0 1"
	var pos PositionNG
	pos.Set(fen)

	states := make([]StateInfo, 12)
	moves := []string{
		"a1e1", // r a1-e1+ (check: e1 attacks e0 via e-file)
		"e0d0", // K e0-d0 (evasion; d0 doesn't face f9)
		"e1d1", // r e1-d1+ (check: d1 attacks d0 via d-file)
		"d0e0", // K d0-e0
		"d1e1", // r d1-e1+ (check) — cycle starts
		"e0d0", // K e0-d0
		"e1d1", // r e1-d1+ (check)
		"d0e0", // K d0-e0 — cycle complete
	}
	applyMoves(t, &pos, moves, states)

	if !pos.IsRepetition() {
		t.Fatal("expected repetition")
	}

	// BLACK to move. Black checked on all its moves → REP_LOSE.
	result := pos.ClassifyRepetition()
	if result != REP_LOSE {
		t.Fatalf("expected REP_LOSE, got %d", result)
	}
}

// TestRepetitionSimpleDraw — neither side checks or chases → REP_DRAW.
func TestRepetitionSimpleDraw(t *testing.T) {
	// White: K f0, R a0.  Black: K d9.
	// Kings on f0 and d9 — different files, no facing.
	// Rook shuttles a0↔a1, black king shuttles d9↔e9.
	fen := "3k5/9/9/9/9/9/9/9/9/R4K3 w - - 0 1"
	var pos PositionNG
	pos.Set(fen)

	states := make([]StateInfo, 12)
	moves := []string{
		"a0a1", "d9e9",
		"a1a0", "e9d9",
		"a0a1", "d9e9",
		"a1a0", "e9d9",
	}
	applyMoves(t, &pos, moves, states)

	if !pos.IsRepetition() {
		t.Fatal("expected repetition")
	}

	result := pos.ClassifyRepetition()
	if result != REP_DRAW {
		t.Fatalf("expected REP_DRAW, got %d", result)
	}
}

// TestRepetitionIsRepetitionBasic — tests cycle detection. A 4-ply cycle
// returns to the starting position, which IS a repetition (offset 4 matches).
func TestRepetitionIsRepetitionBasic(t *testing.T) {
	fen := "3k5/9/9/9/9/9/9/9/9/R4K3 w - - 0 1"
	var pos PositionNG
	pos.Set(fen)

	states := make([]StateInfo, 12)

	// After 2 plies — no repetition (need offset >= 4)
	applyMoves(t, &pos, []string{"a0a1", "d9e9"}, states)
	if pos.IsRepetition() {
		t.Fatal("should NOT detect repetition after only 2 plies")
	}

	// Complete the cycle (4 plies total) — the starting position recurs at offset 4
	applyMoves(t, &pos, []string{"a1a0", "e9d9"}, states[2:])
	if !pos.IsRepetition() {
		t.Fatal("should detect repetition after returning to starting position (4 plies)")
	}
}

// TestRepetitionCheckerLosesFromCheckerPerspective — white perpetually checks,
// white to move → REP_LOSE for white.
func TestRepetitionCheckerLosesFromCheckerPerspective(t *testing.T) {
	// Start: black to move. Black makes innocent king move, then white checks.
	// White: K f0, R a1.  Black: K e9.
	fen := "4k4/9/9/9/9/9/9/9/R8/5K3 b - - 0 1"
	var pos PositionNG
	pos.Set(fen)

	states := make([]StateInfo, 14)
	moves := []string{
		"e9d9", // K e9-d9 (black, innocent; d9 doesn't face f0)
		"a1d1", // R a1-d1+ (white, check via d-file)
		"d9e9", // K d9-e9
		"d1e1", // R d1-e1+ (white, check via e-file)
		"e9d9", // K e9-d9 — cycle starts (matches after ply 1)
		"e1d1", // R e1-d1+ (white, check)
		"d9e9", // K d9-e9
		"d1e1", // R d1-e1+ (white, check)
		"e9d9", // K e9-d9 — cycle complete
	}
	applyMoves(t, &pos, moves, states)

	// 9 plies from B start → W to move. White checked every white move.
	if !pos.IsRepetition() {
		t.Fatal("expected repetition")
	}

	result := pos.ClassifyRepetition()
	if result != REP_LOSE {
		t.Fatalf("expected REP_LOSE for white (checker, to move), got %d", result)
	}
}
