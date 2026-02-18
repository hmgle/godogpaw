package engine

// from Ethereal

type STAGE_T = int8

const (
	STAGE_TABLE STAGE_T = iota
	STAGE_GENERATE_NOISY
	STAGE_NOISY
	STAGE_KILLER_1
	STAGE_KILLER_2
	STAGE_COUNTER_MOVE
	STAGE_GENERATE_QUIET
	STAGE_QUIET
	STAGE_DONE
)

type MovePicker struct {
	SkipQuiets bool
	Stage      STAGE_T
	Split      uint8
	NoisySize  uint8
	QuietSize  uint8

	TableMove   MoveNG
	Killer1     MoveNG
	Killer2     MoveNG
	CounterMove MoveNG
	Moves       [MAX_MOVES]MoveNG
	Values      [MAX_MOVES]Value

	History *HistoryTable
}

func InitalizeMovePicker(mp *MovePicker, skipQuiets bool, tableMove, killer1, killer2, counterMove MoveNG, history *HistoryTable) {
	mp.SkipQuiets = skipQuiets
	mp.Stage = STAGE_TABLE
	mp.Split = 0
	mp.NoisySize = 0
	mp.QuietSize = 0
	mp.TableMove = tableMove
	if killer1 != tableMove {
		mp.Killer1 = killer1
	} else {
		mp.Killer1 = MOVE_NONE
	}
	if killer2 != tableMove {
		mp.Killer2 = killer2
	} else {
		mp.Killer2 = MOVE_NONE
	}
	// Counter-move: skip if it duplicates TT move or a killer
	if counterMove != tableMove && counterMove != killer1 && counterMove != killer2 {
		mp.CounterMove = counterMove
	} else {
		mp.CounterMove = MOVE_NONE
	}
	mp.History = history
}

func EvaluateNoisyMoves(mp *MovePicker, pos *PositionNG) {
	tmpV := [PIECE_TYPE_NB]Value{7, 3, 6, 2, 5, 4, 8, 5}

	// prune removes a stale entry by swapping it with the tail of the noisy move list.
	prune := func(idx int) bool {
		if mp.NoisySize == 0 {
			return false
		}
		last := mp.NoisySize - 1
		mp.Moves[idx] = mp.Moves[last]
		mp.Values[idx] = mp.Values[last]
		mp.NoisySize = last
		return true
	}

	for i := 0; i < int(mp.NoisySize); i++ {
		move := mp.Moves[i]
		from := FromSQ(move)
		to := ToSQ(move)
		fromPiece := pos.Board[from]
		toPiece := pos.Board[to]

		if fromPiece == NO_PIECE || ColorOf(fromPiece) != pos.SideToMove {
			if prune(i) {
				i--
			}
			continue
		}
		if toPiece == NO_PIECE || ColorOf(toPiece) == pos.SideToMove {
			if prune(i) {
				i--
			}
			continue
		}

		fromPieceType := TypeOf(fromPiece)
		if fromPieceType <= 0 {
			if prune(i) {
				i--
			}
			continue
		}

		// Use the standard MVV-LVA
		mp.Values[i] = PieceValue[MG][toPiece] - tmpV[fromPieceType-1]
	}
}

func EvaluateQuietMoves(mp *MovePicker, pos *PositionNG) {
	for i := mp.Split; i < mp.Split+mp.QuietSize; i++ {
		mp.Values[i] = GetHistoryScore(mp.History, mp.Moves[i], pos.SideToMove)
	}
}

func SelectNextMove(mp *MovePicker, pos *PositionNG) MoveNG {
	var bestMove MoveNG
	switch mp.Stage {
	case STAGE_TABLE:
		// Play the table move if it is from this
		// position, also advance to the next stage
		mp.Stage = STAGE_GENERATE_NOISY
		if mp.TableMove != MOVE_NONE && pos.PseudoLegal(mp.TableMove) {
			return mp.TableMove
		}
		fallthrough
	case STAGE_GENERATE_NOISY:
		// Generate all noisy moves and evaluate them. Set up the
		// split in the array to store quiet and noisy moves. Also,
		// this stage is only a helper. Advance to the next one.
		mp.NoisySize = pos.Generate(CAPTURES, mp.Moves[:])
		EvaluateNoisyMoves(mp, pos)
		mp.Split = mp.NoisySize
		mp.Stage = STAGE_NOISY
		fallthrough
	case STAGE_NOISY:
		// Check to see if there are still more noisy moves
		if mp.NoisySize != 0 {
			// Find highest scoring move
			best := 0
			for i := 1; i < int(mp.NoisySize); i++ {
				if mp.Values[i] > mp.Values[best] {
					best = i
				}
			}
			// Save the best move before overwriting it
			bestMove = mp.Moves[best]

			// Reduce effective move list size
			mp.NoisySize -= 1
			mp.Moves[best] = mp.Moves[mp.NoisySize]
			mp.Values[best] = mp.Values[mp.NoisySize]

			// Don't play the TT move twice
			if bestMove == mp.TableMove {
				return SelectNextMove(mp, pos)
			}
			// Don't play the killer moves twice
			if bestMove == mp.Killer1 {
				mp.Killer1 = MOVE_NONE
			}
			if bestMove == mp.Killer2 {
				mp.Killer2 = MOVE_NONE
			}
			if bestMove == mp.CounterMove {
				mp.CounterMove = MOVE_NONE
			}
			return bestMove
		}

		// If we are using this move picker for the quiescence
		// search, we have exhausted all moves already. Otherwise,
		// we should move onto the quiet moves (+ killers)
		if mp.SkipQuiets {
			mp.Stage = STAGE_DONE
			return MOVE_NONE
		} else {
			mp.Stage = STAGE_KILLER_1
		}
		fallthrough
	case STAGE_KILLER_1:
		// Play the killer move if it is from this position.
		mp.Stage = STAGE_KILLER_2
		if IsOKMove(mp.Killer1) && pos.PseudoLegal(mp.Killer1) {
			return mp.Killer1
		}
		fallthrough
	case STAGE_KILLER_2:
		// Play the killer move if it is from this position.
		mp.Stage = STAGE_COUNTER_MOVE
		if IsOKMove(mp.Killer2) && pos.PseudoLegal(mp.Killer2) {
			return mp.Killer2
		}
		fallthrough
	case STAGE_COUNTER_MOVE:
		// Play the counter-move if it is from this position.
		mp.Stage = STAGE_GENERATE_QUIET
		if IsOKMove(mp.CounterMove) && pos.PseudoLegal(mp.CounterMove) {
			return mp.CounterMove
		}
		fallthrough
	case STAGE_GENERATE_QUIET:
		// Generate all quiet moves and evaluate them
		mp.QuietSize = pos.Generate(QUIETS, mp.Moves[mp.Split:])
		EvaluateQuietMoves(mp, pos)
		mp.Stage = STAGE_QUIET
		fallthrough
	case STAGE_QUIET:
		// Check to see if there are still more quiet moves
		if mp.QuietSize != 0 {
			// Find highest scoring move
			best := mp.Split
			for i := 1 + mp.Split; i < mp.Split+mp.QuietSize; i++ {
				if mp.Values[i] > mp.Values[best] {
					best = i
				}
			}
			// Save the best move before overwriting it
			bestMove = mp.Moves[best]

			// Reduce effective move list size
			mp.QuietSize--
			mp.Moves[best] = mp.Moves[mp.Split+mp.QuietSize]
			mp.Values[best] = mp.Values[mp.Split+mp.QuietSize]

			// Don't play a move more than once
			if bestMove == mp.TableMove ||
				bestMove == mp.Killer1 ||
				bestMove == mp.Killer2 ||
				bestMove == mp.CounterMove {
				return SelectNextMove(mp, pos)
			}
			return bestMove
		}
		// If no quiet moves left, advance stages
		mp.Stage = STAGE_DONE
		fallthrough
	case STAGE_DONE:
		return MOVE_NONE
	default:
		return MOVE_NONE
	}
}
