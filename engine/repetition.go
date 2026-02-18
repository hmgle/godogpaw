package engine

// RepetitionType classifies a position repetition per Xiangqi rules.
type RepetitionType int8

const (
	REP_NONE RepetitionType = iota // No repetition
	REP_DRAW                       // Draw (symmetric or neither side offends)
	REP_WIN                        // We win (opponent offends)
	REP_LOSE                       // We lose (we offend)
)

// ClassifyRepetition determines the type of repetition according to Xiangqi rules.
// Must only be called when IsRepetition() is true.
//
// Algorithm:
//  1. Find cycle length by walking back to the first matching Zobrist key
//  2. Phase 1 — Perpetual check: walk the cycle, count each side's check moves
//  3. Phase 2 — Perpetual chase: if no perpetual check, check for new threats
//
// Returns the classification from the side-to-move's perspective.
func (pos *PositionNG) ClassifyRepetition() RepetitionType {
	st := pos.St.Top()

	// Find the cycle length (first matching key stepping back by 2)
	cycleLen := 0
	maxLookback := st.PliesFromNull
	if stackMax := len(pos.St) - 1; stackMax < maxLookback {
		maxLookback = stackMax
	}
	for i := 4; i <= maxLookback; i += 2 {
		if pos.St.PrevCnt(i).key == st.key {
			cycleLen = i
			break
		}
	}
	if cycleLen == 0 {
		return REP_DRAW // shouldn't happen if IsRepetition() was true
	}

	// Phase 1: Perpetual check detection
	// Walk through the cycle and count how many moves each side made were checks.
	// The cycle contains cycleLen plies. Side-to-move made the move at offset 1, 3, 5...
	// Opponent made the move at offset 2, 4, 6...
	//
	// We look at states [current-1], [current-2], ..., [current-cycleLen].
	// State at offset i was reached by a move. If that state has checkersBB set,
	// the move that led to it was a check.
	//
	// Moves at odd offsets (1, 3, 5...) were made by us (current side-to-move).
	// The state at PrevCnt(i) was reached by a move. When i is odd, the move
	// from state[N-i-1] to state[N-i] was made by sideToMove (us).
	// When i is even, the move was made by the opponent.

	ourChecks := 0
	ourMoves := 0
	theirChecks := 0
	theirMoves := 0

	for i := 1; i <= cycleLen; i++ {
		stI := pos.St.PrevCnt(i)
		isCheck := stI.checkersBB.IsNotZero()
		if i%2 == 1 {
			// Odd offset: move was made by us (current side-to-move)
			ourMoves++
			if isCheck {
				ourChecks++
			}
		} else {
			// Even offset: move was made by the opponent
			theirMoves++
			if isCheck {
				theirChecks++
			}
		}
	}

	weCheckEveryMove := ourMoves > 0 && ourChecks == ourMoves
	theyCheckEveryMove := theirMoves > 0 && theirChecks == theirMoves

	if weCheckEveryMove && theyCheckEveryMove {
		return REP_DRAW
	}
	if weCheckEveryMove {
		return REP_LOSE
	}
	if theyCheckEveryMove {
		return REP_WIN
	}

	// Phase 2: Perpetual chase detection
	// We need to check if one side is chasing (creating new threats) on every move.
	//
	// Build board snapshots by walking backward from the current position,
	// undoing moves locally on a [SQUARE_NB]Piece copy. For each move in the
	// cycle, check if it creates a new threat to an enemy piece.

	ourChases := 0
	theirChases := 0

	// Build the "after" board snapshot from the current position
	var boardAfter [SQUARE_NB]Piece
	copy(boardAfter[:], pos.Board[:])

	for i := 1; i <= cycleLen; i++ {
		stI := pos.St.PrevCnt(i)
		move := stI.lastMove
		if !IsOKMove(move) {
			// Null move or invalid — can't classify chase
			continue
		}

		// Build the "before" board by undoing the move locally
		var boardBefore [SQUARE_NB]Piece
		copy(boardBefore[:], boardAfter[:])
		from := FromSQ(move)
		to := ToSQ(move)
		boardBefore[from] = boardBefore[to]
		boardBefore[to] = stI.capturedPiece // restore captured piece (NO_PIECE if none)

		// Determine who made this move
		moverColor := ColorOf(boardAfter[to])

		if detectChase(boardBefore[:], boardAfter[:], move, moverColor) {
			if i%2 == 1 {
				ourChases++
			} else {
				theirChases++
			}
		}

		// Advance: the "after" for the next iteration (going further back) is this "before"
		copy(boardAfter[:], boardBefore[:])
	}

	weChaseEveryMove := ourMoves > 0 && ourChases == ourMoves
	theyChaseEveryMove := theirMoves > 0 && theirChases == theirMoves

	if weChaseEveryMove && theyChaseEveryMove {
		return REP_DRAW
	}
	if weChaseEveryMove {
		return REP_LOSE
	}
	if theyChaseEveryMove {
		return REP_WIN
	}

	// Check combined offenses: if we check on some moves and chase on the rest
	weAlwaysOffend := ourMoves > 0 && (ourChecks+ourChases) == ourMoves
	theyAlwaysOffend := theirMoves > 0 && (theirChecks+theirChases) == theirMoves

	if weAlwaysOffend && theyAlwaysOffend {
		return REP_DRAW
	}
	if weAlwaysOffend {
		return REP_LOSE
	}
	if theyAlwaysOffend {
		return REP_WIN
	}

	return REP_DRAW
}

// boardContext holds pre-computed bitboards for a board snapshot,
// enabling fast attack lookups without a full PositionNG.
type boardContext struct {
	byColor [COLOR_NB]Bitboard
	byType  [PIECE_TYPE_NB]Bitboard
	occ     Bitboard // all pieces
}

func buildBoardContext(board []Piece) boardContext {
	var ctx boardContext
	for sq := SQ_A0; sq <= SQ_I9; sq++ {
		pc := board[sq]
		if pc == NO_PIECE {
			continue
		}
		bb := SquareBB[sq]
		ctx.occ = ctx.occ.Or(bb)
		ctx.byColor[ColorOf(pc)] = ctx.byColor[ColorOf(pc)].Or(bb)
		ctx.byType[TypeOf(pc)] = ctx.byType[TypeOf(pc)].Or(bb)
	}
	return ctx
}

// attackersOfColor returns a bitboard of all pieces of the given color
// that attack the given square, using the board context's occupancy.
func (ctx *boardContext) attackersOfColor(sq Square, c Color) Bitboard {
	colorBB := ctx.byColor[c]
	return (PawnAttacksTo[c][sq].And(ctx.byType[PAWN]).
		Or(AttacksBB(KNIGHT_TO, sq, ctx.occ).And(ctx.byType[KNIGHT])).
		Or(AttacksBB(ROOK, sq, ctx.occ).And(ctx.byType[ROOK])).
		Or(AttacksBB(ROOK, sq, ctx.occ).And(ctx.byType[KING])). // king attacks along rook lines
		Or(AttacksBB(CANNON, sq, ctx.occ).And(ctx.byType[CANNON]))).And(colorBB)
}

// detectChase checks whether a move creates a new threat (chase) against an enemy piece.
// A chase exists when, after the move, the mover attacks an enemy piece that was
// NOT attacked before the move. The victim must be either unprotected or worth more
// than the attacker. Kings and uncrossed pawns are excluded.
func detectChase(boardBefore, boardAfter []Piece, move MoveNG, moverColor Color) bool {
	ctxBefore := buildBoardContext(boardBefore)
	ctxAfter := buildBoardContext(boardAfter)

	enemyColor := notColor(moverColor)
	to := ToSQ(move)

	// Check each enemy piece on the board after the move
	enemies := ctxAfter.byColor[enemyColor]
	for enemies.IsNotZero() {
		sq := PopLsb(&enemies)
		pc := boardAfter[sq]
		pt := TypeOf(pc)

		// Skip kings (that's check, not chase)
		if pt == KING {
			continue
		}

		// Skip pawns that haven't crossed the river
		if pt == PAWN && !hasCrossedRiver(sq, enemyColor) {
			continue
		}

		// Skip the piece on the destination of the move (just-captured piece is gone)
		if sq == to {
			continue
		}

		// Find new attackers: pieces that attack this square after but not before
		attackersAfter := ctxAfter.attackersOfColor(sq, moverColor)
		attackersBefore := ctxBefore.attackersOfColor(sq, moverColor)
		newAttackers := attackersAfter.And(attackersBefore.Not())

		if !newAttackers.IsNotZero() {
			continue
		}

		// Check if any new attacker constitutes a chase
		victimValue := seeValues[pt]
		defenders := ctxAfter.attackersOfColor(sq, enemyColor)
		isProtected := defenders.IsNotZero()

		for newAttackers.IsNotZero() {
			aSq := PopLsb(&newAttackers)
			attackerPt := TypeOf(boardAfter[aSq])
			attackerValue := seeValues[attackerPt]

			// Chase if victim is unprotected OR attacker is worth less than victim
			if !isProtected || attackerValue < victimValue {
				return true
			}
		}
	}
	return false
}

// hasCrossedRiver returns true if a pawn at sq has crossed the river.
// White pawns cross at rank >= 5, black pawns cross at rank <= 4.
func hasCrossedRiver(sq Square, color Color) bool {
	r := RankOf(sq)
	if color == WHITE {
		return r >= RANK_5
	}
	return r <= RANK_4
}
