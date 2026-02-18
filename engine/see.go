package engine

// SEE (Static Exchange Evaluation) estimates the material outcome of a
// sequence of captures on a single square.  It is used to prune losing
// captures in quiescence search and to improve move ordering.
//
// Xiangqi-specific: when a piece is removed from the board during the
// simulated exchange, the cannon screen configuration may change,
// potentially revealing new attackers.  We handle this by recalculating
// attackers through the removed piece's square after each simulated
// capture.

// seeValues maps piece types to simple material values for SEE.
var seeValues = [PIECE_TYPE_NB]Value{
	0,    // NO_PIECE_TYPE
	1200, // ROOK
	200,  // ADVISOR
	600,  // CANNON
	100,  // PAWN
	550,  // KNIGHT
	200,  // BISHOP
	0,    // KING (infinite, but we use 0 — king captures are always last)
}

// SEE returns the static exchange evaluation score for a capture move.
// A positive score means the capture wins material; negative means it loses.
func (pos *PositionNG) SEE(m MoveNG) Value {
	from := FromSQ(m)
	to := ToSQ(m)

	// The piece being captured (initial victim)
	target := pos.PieceOn(to)
	if target == NO_PIECE {
		return 0 // not a capture
	}

	// Build the set of attackers to the target square.
	occupied := pos.PiecesAllColor(ALL_PIECES)
	attackers := pos.AttackersTo(to, occupied)

	// Gain list: gain[0] = value of initial capture, gain[i] = value recaptured at depth i.
	var gain [32]Value
	depth := 0
	gain[depth] = seeValues[TypeOf(target)]

	// The side making the initial capture.
	stm := ColorOf(pos.PieceOn(from))
	movedPt := TypeOf(pos.PieceOn(from))

	for {
		depth++
		// The value of recapturing = previous gain - what we just captured
		gain[depth] = seeValues[movedPt] - gain[depth-1]

		// Pruning: if the score can't improve even if we don't recapture, stop
		if max32(-gain[depth-1], gain[depth]) < 0 {
			break
		}

		// Remove the piece that just captured from occupied/attackers
		occupied = occupied.And(SquareBB[from].Not())
		attackers = attackers.And(SquareBB[from].Not())

		// Recalculate sliding attackers through the vacated square.
		// Rooks and cannons may now have new attack paths.
		attackers = attackers.Or(
			AttacksBB(ROOK, to, occupied).And(pos.PiecesAllColor(ROOK)).And(occupied),
		).Or(
			AttacksBB(CANNON, to, occupied).And(pos.PiecesAllColor(CANNON)).And(occupied),
		)

		// Switch sides
		stm = notColor(stm)

		// Find the least valuable attacker of the current side
		stmAttackers := attackers.And(pos.Pieces(stm))
		if !stmAttackers.IsNotZero() {
			break
		}

		// Pick the least valuable attacker
		movedPt, from = leastValuableAttacker(pos, stmAttackers)
	}

	// Negamax the gain list
	for depth--; depth > 0; depth-- {
		gain[depth-1] = -max32(-gain[depth-1], gain[depth])
	}

	return gain[0]
}

// leastValuableAttacker finds the least valuable piece in the attacker set
// and returns its type and square.
func leastValuableAttacker(pos *PositionNG, attackers Bitboard) (PieceType, Square) {
	// Search in order of increasing value
	for _, pt := range []PieceType{PAWN, ADVISOR, BISHOP, KNIGHT, CANNON, ROOK, KING} {
		b := attackers.And(pos.PiecesAllColor(pt))
		if b.IsNotZero() {
			return pt, Lsb(b)
		}
	}
	return NO_PIECE_TYPE, SQ_NONE
}

func max32(a, b Value) Value {
	if a > b {
		return a
	}
	return b
}

// SEESign returns true if the SEE score is >= threshold.
// This is a quick check used for pruning without computing the full SEE.
func (pos *PositionNG) SEESign(m MoveNG, threshold Value) bool {
	return pos.SEE(m) >= threshold
}
