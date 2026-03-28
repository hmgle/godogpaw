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
	if pos.St != nil && len(pos.St) > 0 {
		phase := pos.St.Top().Phase
		if phase > TotalPhase {
			return TotalPhase
		}
		if phase < 0 {
			return 0
		}
		return phase
	}
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

func phaseContribution(pt PieceType) int {
	switch pt {
	case ROOK:
		return PhaseRook
	case KNIGHT:
		return PhaseKnight
	case CANNON:
		return PhaseCannon
	case ADVISOR:
		return PhaseAdvisor
	case BISHOP:
		return PhaseBishop
	default:
		return 0
	}
}

var attackUnit = [PIECE_TYPE_NB]Value{
	0, // NO_PIECE_TYPE
	9, // ROOK
	1, // ADVISOR
	7, // CANNON
	3, // PAWN
	6, // KNIGHT
	1, // BISHOP
	0, // KING
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

func pstIndex(piece Piece, sq Square) Square {
	if ColorOf(piece) == BLACK {
		return flipSquare(sq)
	}
	return sq
}

func crossedRiver(c Color, sq Square) bool {
	if c == WHITE {
		return RankOf(sq) >= RANK_5
	}
	return RankOf(sq) <= RANK_4
}

func relativeRank(c Color, sq Square) int {
	if c == WHITE {
		return int(RankOf(sq))
	}
	return int(RANK_9 - RankOf(sq))
}

func forwardSquare(c Color, sq Square) (Square, bool) {
	switch c {
	case WHITE:
		if RankOf(sq) == RANK_9 {
			return SQ_NONE, false
		}
		return sq + NORTH, true
	default:
		if RankOf(sq) == RANK_0 {
			return SQ_NONE, false
		}
		return sq + SOUTH, true
	}
}

func kingStartSquare(c Color) Square {
	if c == WHITE {
		return SQ_E0
	}
	return SQ_E9
}

func kingZone(c Color, kingSq Square) Bitboard {
	r0 := int(RankOf(kingSq))
	f0 := int(FileOf(kingSq))
	zone := From64(0)
	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			r := r0 + dr
			f := f0 + df
			if r < int(RANK_0) || r > int(RANK_9) || f < int(FILE_A) || f > int(FILE_I) {
				continue
			}
			zone = zone.Or(SquareBB[MakeSquareNG(File(f), Rank(r))])
		}
	}
	for df := -1; df <= 1; df++ {
		f := f0 + df
		if f < int(FILE_A) || f > int(FILE_I) {
			continue
		}
		if c == WHITE && r0+2 <= int(RANK_9) {
			zone = zone.Or(SquareBB[MakeSquareNG(File(f), Rank(r0+2))])
		}
		if c == BLACK && r0-2 >= int(RANK_0) {
			zone = zone.Or(SquareBB[MakeSquareNG(File(f), Rank(r0-2))])
		}
	}
	return zone
}

func fileBitboard(f File) Bitboard {
	switch f {
	case FILE_A:
		return FileABB
	case FILE_B:
		return FileBBB
	case FILE_C:
		return FileCBB
	case FILE_D:
		return FileDBB
	case FILE_E:
		return FileEBB
	case FILE_F:
		return FileFBB
	case FILE_G:
		return FileGBB
	case FILE_H:
		return FileHBB
	default:
		return FileIBB
	}
}

func attackWeight(pos *PositionNG, attackers Bitboard) Value {
	var units Value
	for attackers.IsNotZero() {
		sq := PopLsb(&attackers)
		units += attackUnit[TypeOf(pos.PieceOn(sq))]
	}
	return units
}

func leastAttackerValue(pos *PositionNG, attackers Bitboard) Value {
	if !attackers.IsNotZero() {
		return VALUE_ZERO
	}
	pt, _ := leastValuableAttacker(pos, attackers)
	if pt == NO_PIECE_TYPE {
		return VALUE_ZERO
	}
	return PieceValue[MG][MakePieceNG(WHITE, pt)]
}

func (pos *PositionNG) recomputeEvalBase() (int, [COLOR_NB]Value, [COLOR_NB]Value) {
	phase := 0
	var mgBase, egBase [COLOR_NB]Value
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
		idx := pstIndex(piece, sq)
		if pt != KING {
			mgBase[color] += PieceValue[MG][piece]
			egBase[color] += PieceValue[EG][piece]
			phase += phaseContribution(pt)
		}
		mgBase[color] += pstMG[pt][idx]
		egBase[color] += pstEG[pt][idx]
	}
	if phase > TotalPhase {
		phase = TotalPhase
	}
	return phase, mgBase, egBase
}

func (pos *PositionNG) addPawnStructureTerms(mgScore, egScore *[COLOR_NB]Value, occupied Bitboard) {
	for c := Color(WHITE); c < COLOR_NB; c++ {
		opp := notColor(c)
		oppZone := kingZone(opp, pos.KingSQ[opp])
		pawns := pos.Pieces(c, PAWN)
		for pawns.IsNotZero() {
			sq := PopLsb(&pawns)
			relRank := relativeRank(c, sq)
			advanced := max(relRank-4, 0)
			if crossedRiver(c, sq) {
				mgScore[c] += 18 + Value(advanced*4)
				egScore[c] += 28 + Value(advanced*7)
			}
			file := FileOf(sq)
			if file >= FILE_D && file <= FILE_F {
				mgScore[c] += 6
				egScore[c] += 4
			}
			if relRank >= 8 {
				mgScore[c] += 10
				egScore[c] += 18
			}

			support := PawnAttacksTo[c][sq].And(pos.Pieces(c, PAWN))
			if support.IsNotZero() {
				count := Value(support.PopCount())
				mgScore[c] += 8 * count
				egScore[c] += 10 * count
			}
			if file > FILE_A && pos.PieceOn(MakeSquareNG(file-1, RankOf(sq))) == MakePieceNG(c, PAWN) {
				mgScore[c] += 4
				egScore[c] += 4
			}
			if file < FILE_I && pos.PieceOn(MakeSquareNG(file+1, RankOf(sq))) == MakePieceNG(c, PAWN) {
				mgScore[c] += 4
				egScore[c] += 4
			}

			if frontSq, ok := forwardSquare(c, sq); ok && !crossedRiver(c, sq) && pos.PieceOn(frontSq) != NO_PIECE {
				mgScore[c] -= 8
				egScore[c] -= 4
			}

			pressure := PawnAttacks[c][sq].And(oppZone).PopCount()
			if pressure > 0 {
				mgScore[c] += Value(pressure) * 10
				egScore[c] += Value(pressure) * 8
			}
		}
	}
}

func (pos *PositionNG) addPieceActivityTerms(mgScore, egScore *[COLOR_NB]Value, occupied Bitboard) {
	for c := Color(WHITE); c < COLOR_NB; c++ {
		opp := notColor(c)
		oppKingSq := pos.KingSQ[opp]

		rookBB := pos.Pieces(c, ROOK)
		for rookBB.IsNotZero() {
			sq := PopLsb(&rookBB)
			mobility := Value(AttacksBB(ROOK, sq, occupied).PopCount())
			mgScore[c] += mobility * 3
			egScore[c] += mobility * 4

			fileOcc := occupied.And(fileBitboard(FileOf(sq)))
			ownPawns := pos.Pieces(c, PAWN).And(fileOcc).PopCount()
			oppPawns := pos.Pieces(opp, PAWN).And(fileOcc).PopCount()
			if ownPawns == 0 && oppPawns == 0 {
				mgScore[c] += 18
				egScore[c] += 14
			} else if ownPawns == 0 {
				mgScore[c] += 10
				egScore[c] += 8
			}
			if FileOf(sq) == FileOf(oppKingSq) {
				between := BetweenBB[sq][oppKingSq].And(occupied).PopCount()
				if between == 0 {
					mgScore[c] += 28
					egScore[c] += 18
				} else if between == 1 {
					mgScore[c] += 14
					egScore[c] += 8
				}
			}
		}

		knightBB := pos.Pieces(c, KNIGHT)
		for knightBB.IsNotZero() {
			sq := PopLsb(&knightBB)
			rawMobility := Value(AttacksBBEmptyOcc(KNIGHT, sq).PopCount())
			mobility := Value(AttacksBB(KNIGHT, sq, occupied).PopCount())
			blocked := rawMobility - mobility
			mgScore[c] += mobility * 4
			egScore[c] += mobility * 5
			mgScore[c] -= blocked * 6
			egScore[c] -= blocked * 3
			if Distance(sq, oppKingSq) <= 2 {
				mgScore[c] += 18
				egScore[c] += 12
			}
		}

		cannonBB := pos.Pieces(c, CANNON)
		for cannonBB.IsNotZero() {
			sq := PopLsb(&cannonBB)
			mobility := Value(AttacksBB(CANNON, sq, occupied).PopCount())
			mgScore[c] += mobility * 2
			egScore[c] += mobility * 2

			if LineBB[sq][oppKingSq].IsNotZero() {
				screens := BetweenBB[sq][oppKingSq].And(occupied).PopCount()
				switch screens {
				case 1:
					mgScore[c] += 36
					egScore[c] += 18
				case 2:
					mgScore[c] += 10
				case 0:
					mgScore[c] += 6
				}
			}
		}
	}
}

func (pos *PositionNG) addKingSafetyTerms(mgScore, egScore *[COLOR_NB]Value, occupied Bitboard) {
	for c := Color(WHITE); c < COLOR_NB; c++ {
		opp := notColor(c)
		kingSq := pos.KingSQ[c]
		zone := kingZone(c, kingSq)
		advisorCount := pos.PieceCount[MakePieceNG(c, ADVISOR)]
		bishopCount := pos.PieceCount[MakePieceNG(c, BISHOP)]

		if advisorCount == 2 && bishopCount == 2 {
			mgScore[c] += 32
			egScore[c] += 22
		}
		mgScore[c] += Value(advisorCount * 10)
		egScore[c] += Value(advisorCount * 6)
		mgScore[c] += Value(bishopCount * 8)
		egScore[c] += Value(bishopCount * 6)

		if kingSq == kingStartSquare(c) {
			mgScore[c] += 12
		}
		if FileOf(kingSq) != FILE_E {
			mgScore[c] -= 10
		}
		if relativeRank(c, kingSq) > 1 {
			mgScore[c] -= 16
			egScore[c] -= 6
		}

		weakness := Value(16 - 3*advisorCount - 2*bishopCount)
		if weakness < 4 {
			weakness = 4
		}
		enemyPressure := Value(0)
		undefended := Value(0)
		for sq := Square(0); sq < SQUARE_NB; sq++ {
			if !zone.And(SquareBB[sq]).IsNotZero() {
				continue
			}
			enemyAttackers := pos.AttackersTo(sq, occupied).And(pos.Pieces(opp))
			if !enemyAttackers.IsNotZero() {
				continue
			}
			friendlyDefenders := pos.AttackersTo(sq, occupied).And(pos.Pieces(c))
			enemyPressure += attackWeight(pos, enemyAttackers)
			if !friendlyDefenders.IsNotZero() {
				undefended++
			}
		}
		mgScore[c] -= enemyPressure * weakness / 5
		egScore[c] -= enemyPressure * (weakness + 2) / 8
		mgScore[c] -= undefended * 12
		egScore[c] -= undefended * 6

		oppRookBB := pos.Pieces(opp, ROOK)
		for oppRookBB.IsNotZero() {
			rSq := PopLsb(&oppRookBB)
			if !LineBB[rSq][kingSq].IsNotZero() {
				continue
			}
			between := BetweenBB[rSq][kingSq].And(occupied).PopCount()
			if between == 0 {
				mgScore[c] -= 48
				egScore[c] -= 26
			} else if between == 1 {
				mgScore[c] -= 18
				egScore[c] -= 10
			}
		}

		oppCannonBB := pos.Pieces(opp, CANNON)
		for oppCannonBB.IsNotZero() {
			rSq := PopLsb(&oppCannonBB)
			if !LineBB[rSq][kingSq].IsNotZero() {
				continue
			}
			between := BetweenBB[rSq][kingSq].And(occupied).PopCount()
			if between == 1 {
				mgScore[c] -= 34
				egScore[c] -= 14
			} else if between == 2 {
				mgScore[c] -= 10
			}
		}
	}
}

func (pos *PositionNG) addThreatTerms(mgScore, egScore *[COLOR_NB]Value, occupied Bitboard) {
	for c := Color(WHITE); c < COLOR_NB; c++ {
		opp := notColor(c)
		enemy := pos.Pieces(opp)
		for enemy.IsNotZero() {
			sq := PopLsb(&enemy)
			piece := pos.PieceOn(sq)
			if piece == NO_PIECE || TypeOf(piece) == KING {
				continue
			}
			victimValue := PieceValue[MG][piece]
			attackers := pos.AttackersTo(sq, occupied).And(pos.Pieces(c))
			if !attackers.IsNotZero() {
				continue
			}
			defenders := pos.AttackersTo(sq, occupied).And(pos.Pieces(opp))
			cheapest := leastAttackerValue(pos, attackers)

			if !defenders.IsNotZero() {
				mgScore[c] += victimValue / 7
				egScore[c] += victimValue / 9
				continue
			}
			attackCount := Value(attackers.PopCount())
			defendCount := Value(defenders.PopCount())
			if cheapest > 0 && cheapest < victimValue {
				gain := victimValue - cheapest
				mgScore[c] += gain/10 + attackCount*4
				egScore[c] += gain/12 + attackCount*3
			}
			if attackCount > defendCount {
				mgScore[c] += victimValue / 14
				egScore[c] += victimValue / 16
			}
		}
	}
}

func (pos *PositionNG) evaluateWithBase(phase int, mgBase, egBase [COLOR_NB]Value) Value {
	mgScore := mgBase
	egScore := egBase
	occupied := pos.PiecesAllColor(ALL_PIECES)

	pos.addPawnStructureTerms(&mgScore, &egScore, occupied)
	pos.addPieceActivityTerms(&mgScore, &egScore, occupied)
	pos.addKingSafetyTerms(&mgScore, &egScore, occupied)
	pos.addThreatTerms(&mgScore, &egScore, occupied)

	mg := mgScore[WHITE] - mgScore[BLACK]
	eg := egScore[WHITE] - egScore[BLACK]
	score := (mg*Value(phase) + eg*Value(TotalPhase-phase)) / Value(TotalPhase)
	score += 3
	if pos.SideToMove == BLACK {
		score = -score
	}
	return score
}

func (pos *PositionNG) evaluateNoCache() Value {
	phase, mgBase, egBase := pos.recomputeEvalBase()
	return pos.evaluateWithBase(phase, mgBase, egBase)
}

func (pos *PositionNG) Evaluate() Value {
	if pos.St == nil || len(pos.St) == 0 {
		return pos.evaluateNoCache()
	}
	st := pos.St.Top()
	var mgBase, egBase [COLOR_NB]Value
	for c := Color(WHITE); c < COLOR_NB; c++ {
		mgBase[c] = st.Material[c] + st.PST[MG][c]
		egBase[c] = st.MaterialEG[c] + st.PST[EG][c]
	}
	return pos.evaluateWithBase(pos.gamePhase(), mgBase, egBase)
}
