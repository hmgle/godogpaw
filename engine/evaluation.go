package engine

// Phase weights for game phase calculation.
// Higher values mean the piece contributes more to "midgame-ness".
const (
	PhaseRook    = 6
	PhaseKnight  = 3
	PhaseCannon  = 3
	PhaseAdvisor = 1
	PhaseBishop  = 1
	TotalPhase   = 2 * (2*PhaseRook + 2*PhaseKnight + 2*PhaseCannon + 2*PhaseAdvisor + 2*PhaseBishop) // 56
)

// gamePhase returns the current game phase (0 = endgame, TotalPhase = midgame).
func (pos *PositionNG) gamePhase() int {
	phase := 0
	phase += pos.PieceCount[W_ROOK] * PhaseRook
	phase += pos.PieceCount[B_ROOK] * PhaseRook
	phase += pos.PieceCount[W_KNIGHT] * PhaseKnight
	phase += pos.PieceCount[B_KNIGHT] * PhaseKnight
	phase += pos.PieceCount[W_CANNON] * PhaseCannon
	phase += pos.PieceCount[B_CANNON] * PhaseCannon
	phase += pos.PieceCount[W_ADVISOR] * PhaseAdvisor
	phase += pos.PieceCount[B_ADVISOR] * PhaseAdvisor
	phase += pos.PieceCount[W_BISHOP] * PhaseBishop
	phase += pos.PieceCount[B_BISHOP] * PhaseBishop
	if phase > TotalPhase {
		phase = TotalPhase
	}
	return phase
}

// Midgame piece-square tables (positional bonus, from White's perspective).
// Rank 0 = White's back rank, Rank 9 = opponent's back rank.
// Array index = rank*9 + file.
var pstMG = [PIECE_TYPE_NB][SQUARE_NB]Value{
	{}, // NO_PIECE_TYPE
	{ // ROOK MG - prefers central files and advanced positions
		-6, -4, -2, 0, 6, 0, -2, -4, -6,
		-4, 0, 4, 8, 8, 8, 4, 0, -4,
		-2, 2, 6, 8, 10, 8, 6, 2, -2,
		0, 4, 8, 10, 12, 10, 8, 4, 0,
		4, 8, 10, 12, 14, 12, 10, 8, 4,
		4, 8, 12, 14, 16, 14, 12, 8, 4,
		8, 12, 16, 18, 20, 18, 16, 12, 8,
		8, 10, 14, 18, 22, 18, 14, 10, 8,
		10, 14, 18, 22, 30, 22, 18, 14, 10,
		8, 10, 12, 16, 18, 16, 12, 10, 8,
	},
	{ // ADVISOR MG - only 5 valid positions in palace
		// Valid: d0(3), f0(5), e1(13), d2(21), f2(23)
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 15, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
	{ // CANNON MG - center (e-file) is strong, prefers own half with screens
		0, 0, 2, 4, 8, 4, 2, 0, 0,
		0, 2, 4, 6, 12, 6, 4, 2, 0,
		2, 4, 6, 8, 14, 8, 6, 4, 2,
		0, 2, 4, 6, 10, 6, 4, 2, 0,
		-2, 0, 2, 4, 8, 4, 2, 0, -2,
		-2, -2, 0, 2, 6, 2, 0, -2, -2,
		-4, -2, -2, 0, 4, 0, -2, -2, -4,
		-6, -4, -4, -2, 0, -2, -4, -4, -6,
		-6, -4, -4, -2, 0, -2, -4, -4, -6,
		-8, -6, -6, -4, -2, -4, -6, -6, -8,
	},
	{ // PAWN MG - big bonus for crossing river, throat position (rank 7 center) highest
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 4, 0, 0, 0, 0,
		2, 0, 4, 0, 8, 0, 4, 0, 2,
		6, 12, 18, 22, 28, 22, 18, 12, 6,
		10, 20, 28, 34, 40, 34, 28, 20, 10,
		14, 26, 36, 46, 52, 46, 36, 26, 14,
		10, 20, 30, 44, 50, 44, 30, 20, 10,
		2, 6, 10, 18, 20, 18, 10, 6, 2,
	},
	{ // KNIGHT MG - center and advanced positions preferred, rim is poor
		-10, 0, -4, -2, 0, -2, -4, 0, -10,
		-6, 0, 0, 2, 0, 2, 0, 0, -6,
		-2, 4, 6, 8, 4, 8, 6, 4, -2,
		0, 6, 10, 10, 12, 10, 10, 6, 0,
		4, 8, 12, 14, 16, 14, 12, 8, 4,
		4, 10, 14, 18, 20, 18, 14, 10, 4,
		2, 12, 16, 22, 24, 22, 16, 12, 2,
		0, 8, 14, 20, 22, 20, 14, 8, 0,
		-4, 4, 10, 18, 20, 18, 10, 4, -4,
		-10, -2, 0, 8, 10, 8, 0, -2, -10,
	},
	{ // BISHOP MG - only 7 valid positions for white: c0,g0,a2,e2,i2,c4,g4
		// Center bishop (e2) has 4 moves and is strongest
		0, 0, -2, 0, 0, 0, -2, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		-5, 0, 0, 0, 10, 0, 0, 0, -5,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 5, 0, 0, 0, 5, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
	{ // KING MG - prefer starting square e0, penalize leaving back rank
		// Valid: d0-f0, d1-f1, d2-f2
		0, 0, 0, 2, 10, 2, 0, 0, 0,
		0, 0, 0, -2, -5, -2, 0, 0, 0,
		0, 0, 0, -8, -10, -8, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// Endgame piece-square tables.
var pstEG = [PIECE_TYPE_NB][SQUARE_NB]Value{
	{}, // NO_PIECE_TYPE
	{ // ROOK EG - activity matters more, more uniform
		-4, -2, 0, 2, 4, 2, 0, -2, -4,
		-2, 2, 4, 6, 6, 6, 4, 2, -2,
		0, 4, 6, 8, 8, 8, 6, 4, 0,
		2, 6, 8, 10, 10, 10, 8, 6, 2,
		4, 8, 10, 12, 12, 12, 10, 8, 4,
		6, 10, 12, 14, 14, 14, 12, 10, 6,
		8, 12, 14, 16, 18, 16, 14, 12, 8,
		8, 12, 14, 18, 20, 18, 14, 12, 8,
		10, 14, 16, 20, 24, 20, 16, 14, 10,
		8, 10, 12, 16, 18, 16, 12, 10, 8,
	},
	{ // ADVISOR EG
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 10, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
	{ // CANNON EG - weaker in endgame (fewer screens)
		-4, -2, -2, 0, 2, 0, -2, -2, -4,
		-2, 0, 0, 2, 4, 2, 0, 0, -2,
		-2, 0, 2, 4, 6, 4, 2, 0, -2,
		-2, 0, 2, 4, 4, 4, 2, 0, -2,
		-4, -2, 0, 2, 4, 2, 0, -2, -4,
		-4, -2, -2, 0, 2, 0, -2, -2, -4,
		-4, -4, -2, 0, 0, 0, -2, -4, -4,
		-6, -4, -4, -2, -2, -2, -4, -4, -6,
		-6, -4, -4, -4, -2, -4, -4, -4, -6,
		-8, -6, -6, -4, -4, -4, -6, -6, -8,
	},
	{ // PAWN EG - crossed pawns are even more valuable in endgame
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 6, 0, 0, 0, 0,
		4, 0, 6, 0, 10, 0, 6, 0, 4,
		10, 16, 24, 30, 36, 30, 24, 16, 10,
		16, 28, 36, 44, 52, 44, 36, 28, 16,
		20, 34, 46, 56, 64, 56, 46, 34, 20,
		16, 28, 40, 54, 62, 54, 40, 28, 16,
		4, 10, 16, 24, 28, 24, 16, 10, 4,
	},
	{ // KNIGHT EG - stronger in endgame, prefers close to opponent king
		-8, 0, -2, 0, 2, 0, -2, 0, -8,
		-4, 2, 4, 6, 4, 6, 4, 2, -4,
		0, 6, 8, 10, 8, 10, 8, 6, 0,
		2, 8, 12, 14, 14, 14, 12, 8, 2,
		4, 10, 14, 18, 18, 18, 14, 10, 4,
		6, 12, 16, 20, 22, 20, 16, 12, 6,
		4, 14, 18, 24, 28, 24, 18, 14, 4,
		2, 10, 16, 22, 26, 22, 16, 10, 2,
		-2, 6, 12, 20, 24, 20, 12, 6, -2,
		-6, 0, 4, 10, 14, 10, 4, 0, -6,
	},
	{ // BISHOP EG
		0, 0, -4, 0, 0, 0, -4, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		-5, 0, 0, 0, 5, 0, 0, 0, -5,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
	{ // KING EG - more active king needed
		0, 0, 0, -2, 0, -2, 0, 0, 0,
		0, 0, 0, 0, 2, 0, 0, 0, 0,
		0, 0, 0, 0, 2, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
	},
}

func flipSquare(sq Square) Square {
	return MakeSquareNG(FileOf(sq), RANK_9-RankOf(sq))
}

func (pos *PositionNG) Evaluate() Value {
	phase := pos.gamePhase()

	var mgScore, egScore [COLOR_NB]Value

	occupied := pos.PiecesAllColor(ALL_PIECES)

	// Material + PST
	for sq := Square(0); sq < SQUARE_NB; sq++ {
		piece := pos.Board[sq]
		if piece == NO_PIECE {
			continue
		}
		pt := TypeOf(piece)
		if pt <= NO_PIECE_TYPE || pt >= PIECE_TYPE_NB {
			continue
		}
		color := ColorOf(piece)
		idx := sq
		if color == BLACK {
			idx = flipSquare(sq)
		}
		mgScore[color] += PieceValue[MG][piece] + pstMG[pt][idx]
		egScore[color] += PieceValue[EG][piece] + pstEG[pt][idx]
	}

	// Evaluation terms per side
	for c := Color(WHITE); c < COLOR_NB; c++ {
		opp := notColor(c)

		// --- Advisor/Bishop completeness ---
		advisorCount := pos.PieceCount[MakePieceNG(c, ADVISOR)]
		bishopCount := pos.PieceCount[MakePieceNG(c, BISHOP)]
		oppRookCount := pos.PieceCount[MakePieceNG(opp, ROOK)]

		// Full defensive structure bonus
		if advisorCount == 2 && bishopCount == 2 {
			mgScore[c] += 40
			egScore[c] += 30
		}
		// Advisor penalties
		if advisorCount == 0 {
			penalty := Value(30)
			if oppRookCount > 0 {
				penalty = 60
			}
			mgScore[c] -= penalty
			egScore[c] -= penalty
		} else if advisorCount == 1 && oppRookCount > 0 {
			mgScore[c] -= 30
			egScore[c] -= 20
		}
		// Bishop penalties
		if bishopCount == 0 {
			penalty := Value(25)
			if oppRookCount > 0 {
				penalty = 50
			}
			mgScore[c] -= penalty
			egScore[c] -= penalty
		} else if bishopCount == 1 && oppRookCount > 0 {
			mgScore[c] -= 20
			egScore[c] -= 15
		}

		// --- Rook mobility ---
		rookBB := pos.Pieces(c, ROOK)
		for rookBB.IsNotZero() {
			sq := PopLsb(&rookBB)
			mobility := int(AttacksBB(ROOK, sq, occupied).PopCount())
			mgScore[c] += Value(mobility * 2)
			egScore[c] += Value(mobility * 3)
		}

		// --- Knight trapped detection ---
		knightBB := pos.Pieces(c, KNIGHT)
		for knightBB.IsNotZero() {
			sq := PopLsb(&knightBB)
			mobility := int(AttacksBB(KNIGHT, sq, occupied).PopCount())
			if mobility == 0 {
				mgScore[c] -= 30
				egScore[c] -= 30
			} else if mobility == 1 {
				mgScore[c] -= 15
				egScore[c] -= 15
			}
		}

		// --- Cannon evaluation ---
		cannonBB := pos.Pieces(c, CANNON)
		oppKingSq := pos.KingSQ[opp]
		for cannonBB.IsNotZero() {
			sq := PopLsb(&cannonBB)
			// Cannon on same file as opponent king = pressure
			if FileOf(sq) == FileOf(oppKingSq) {
				mgScore[c] += 15
				// Hollow cannon: no pieces between cannon and king on the file
				between := BetweenBB[sq][oppKingSq].And(SquareBB[oppKingSq].Not()).And(occupied)
				if !between.IsNotZero() {
					mgScore[c] += 20
					egScore[c] += 10
				}
			}
		}

		// --- King safety ---
		kingSq := pos.KingSQ[c]
		// King on starting square bonus
		startSq := SQ_E0
		if c == BLACK {
			startSq = SQ_E9
		}
		if kingSq == startSq {
			mgScore[c] += 10
		}
		// Opponent rook on same file as our king
		oppRookBB := pos.Pieces(opp, ROOK)
		for oppRookBB.IsNotZero() {
			rSq := PopLsb(&oppRookBB)
			if FileOf(rSq) == FileOf(kingSq) {
				mgScore[c] -= 20
				egScore[c] -= 15
			}
		}
	}

	// Interpolate MG/EG scores using game phase
	mg := mgScore[WHITE] - mgScore[BLACK]
	eg := egScore[WHITE] - egScore[BLACK]
	score := (mg*Value(phase) + eg*Value(TotalPhase-phase)) / Value(TotalPhase)

	// Tempo bonus
	score += 3

	if pos.SideToMove == BLACK {
		score = -score
	}
	return score
}
