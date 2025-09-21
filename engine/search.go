package engine

import (
	"fmt"
	"time"
)

var (
	PvTable  = [MAX_MOVES * MAX_MOVES]MoveNG{}
	PvLength = [MAX_MOVES]int{}
)

func StorePvMove(move MoveNG, searchPly int) {
	PvTable[searchPly*int(MAX_MOVES)+searchPly] = move
	for nextPly := searchPly + 1; nextPly < PvLength[searchPly+1]; nextPly++ {
		PvTable[searchPly*int(MAX_MOVES)+nextPly] = PvTable[(searchPly+1)*int(MAX_MOVES)+nextPly]
	}
	PvLength[searchPly] = PvLength[searchPly+1]
}

func Quiescence(alpha, beta Value, pos *PositionNG) (bestScore Value) {
	PvLength[pos.GamePly] = pos.GamePly
	evalation := pos.Evaluate()
	if pos.GamePly >= int(MAX_MOVES) {
		return evalation
	}
	if evalation >= beta {
		return evalation
	}
	if evalation > alpha {
		alpha = evalation
	}

	var mp MovePicker
	InitalizeMovePicker(&mp, true, MOVE_NONE, MOVE_NONE, MOVE_NONE, &pos.History)
	for currentMove := SelectNextMove(&mp, pos); currentMove != MOVE_NONE; currentMove = SelectNextMove(&mp, pos) {
		if !pos.Legal(currentMove) {
			continue
		}
		var st StateInfo
		pos.DoMove(currentMove, &st)
		score := -Quiescence(-beta, -alpha, pos)
		/*
			StorePvMove: a:-38, b: -37, score(0)): -37, ply: 4, move: i0h0, pos:
			-Quiescence(37, 38)
			eval = -34*2 = -68
		*/
		pos.UndoMove(currentMove)
		if score > alpha {
			// Update the Principle Variation
			// fmt.Printf("xxx StorePvMove: %s, score: %d, searchPly: %d\n", pos.MoveStr(currentMove), score, pos.GamePly)
			// fmt.Printf("StorePvMove: a:%d, b: %d, score(%d)): %v, ply: %d, move: %s, pos: %s\n",
			// 	alpha, beta, pos.SideToMove, score, pos.GamePly, Move2Str(currentMove), pos.String())
			StorePvMove(currentMove, pos.GamePly)
			alpha = score

			if score >= beta {
				// log.Printf("xxxxxxxxxxxx==============xxx\n")
				// time.Sleep(time.Second * 5)
				return score
			}
		}
	}
	return alpha
}

func Negamax(alpha, beta Value, pos *PositionNG, depth uint8, doNullMove bool) (bestScore Value) {
	PvLength[pos.GamePly] = pos.GamePly
	rootNode := pos.GamePly == 0
	pvNode := alpha != beta-1
	hashFlag := TT_ALPHA
	var score Value
	var legalMoves int
	futilityPruning := 0
	if pos.IsDraw() {
		return 0
	}
	var ttMove MoveNG
	var bestMove MoveNG
	if pos.GamePly > 0 {
		var scoreInt16 int16
		scoreInt16, ttMove = readHashEntry(pos.St.Top().key, int16(alpha), int16(beta), &bestMove, depth, uint8(pos.GamePly))
		score = int32(scoreInt16)
		if score != int32(NO_HASH) && !pvNode {
			return score
		}
	}
	if pos.GamePly > 0 && pos.IsRepetition() {
		return -MATERIAL_WEIGHTS[W_CANNON]
	}
	if depth == 0 {
		return Quiescence(alpha, beta, pos)
	}

	// mate distance pruning
	if alpha < -int32(MATE_VALUE) {
		alpha = -int32(MATE_VALUE)
	}
	if beta > int32(MATE_VALUE-1) {
		beta = int32(MATE_VALUE) - 1
	}
	if alpha >= beta {
		return alpha
	}
	inCheck := pos.Checkers().IsNotZero()
	if inCheck {
		depth++
	}

	var staticEval Value
	staticEvalReady := false
	if !pvNode && !inCheck {
		staticEval = pos.Evaluate()
		staticEvalReady = true
		if depth <= 5 && !rootNode && beta > -1000 && alpha < 1000 {
			if staticEval < alpha-int32(depth)*200 { // fail-low
				return staticEval
			}
			if staticEval > beta+int32(depth)*125 { // fali-high
				return staticEval
			}
		}
		if doNullMove {
			if pos.GamePly > 0 && depth > 2 && staticEval >= beta {
				var st StateInfo
				pos.DoNullMove(&st)
				score = -Negamax(-beta, -beta+1, pos, depth-1-2, false)
				pos.UndoNullMove()

				if score >= beta {
					return beta
				}
			}

			// razoring
			score = staticEval + MATERIAL_WEIGHTS[W_PAWN]

			var newScore Value
			if score < beta {
				if depth == 1 {
					newScore = Quiescence(alpha, beta, pos)
					if newScore > score {
						return newScore
					}
					return score
				}
			}
			score += MATERIAL_WEIGHTS[W_PAWN]

			if score < beta && depth < 4 {
				newScore = Quiescence(alpha, beta, pos)
				if newScore < beta {
					if newScore > score {
						return newScore
					}
					return score
				}
			}
		}

		// futility pruning condition
		if !staticEvalReady {
			staticEval = pos.Evaluate()
			staticEvalReady = true
		}
		futilityMargin := [...]Value{0, MATERIAL_WEIGHTS[W_PAWN], MATERIAL_WEIGHTS[W_KNIGHT], MATERIAL_WEIGHTS[W_CANNON]}
		if depth < 4 && abs(int(alpha)) < int(MATE_SCORE) && staticEval+futilityMargin[depth] <= alpha {
			futilityPruning = 1
		}
	}

	movesSearched := 0
	// loop over moves
	var mp MovePicker
	InitalizeMovePicker(&mp, false, ttMove, pos.Killers[pos.GamePly][0], pos.Killers[pos.GamePly][1], &pos.History)
	for currentMove := SelectNextMove(&mp, pos); currentMove != MOVE_NONE; currentMove = SelectNextMove(&mp, pos) {
		if !pos.Legal(currentMove) {
			continue
		}
		legalMoves++

		// futility pruning
		if futilityPruning > 0 && movesSearched > 0 && !pos.Capture(currentMove) && !pos.GivesCheck(currentMove) {
			continue
		}
		var st StateInfo
		pos.DoMove(currentMove, &st)
		if depth < 5 || movesSearched == 0 {
			score = -Negamax(-beta, -alpha, pos, depth-1, true)
		} else {
			// LMR
			if IsOKMove(pos.Killers[pos.GamePly][0]) && IsOKMove(pos.Killers[pos.GamePly][1]) {
				mFrom := FromSQ(currentMove)
				mTo := ToSQ(currentMove)
				k0From := FromSQ(pos.Killers[pos.GamePly][0])
				k0To := ToSQ(pos.Killers[pos.GamePly][0])
				k1From := FromSQ(pos.Killers[pos.GamePly][1])
				k1To := ToSQ(pos.Killers[pos.GamePly][1])
				if !pvNode && movesSearched > 3 && depth > 2 &&
					!inCheck &&
					(mFrom != k0From || mTo != k0To) &&
					(mFrom != k1From || mTo != k1To) &&
					!pos.Capture(currentMove) {
					score = -Negamax(-alpha-1, -alpha, pos, depth-2, true)
				} else {
					score = alpha + 1
				}
			} else {
				score = alpha + 1
			}
			// PVS
			if score > alpha {
				score = -Negamax(-alpha-1, -alpha, pos, depth-1, true)
				if (score > alpha) && score < beta {
					score = -Negamax(-beta, -alpha, pos, depth-1, true)
				}
			}
		}

		pos.UndoMove(currentMove)
		movesSearched++

		if score > alpha {
			hashFlag = TT_EXACT
			bestMove = currentMove
			alpha = score
			// fmt.Printf("vvv StorePvMove: %s, score: %d, depth: %d, alpha: %d, searchPly: %d\n", pos.MoveStr(currentMove), score, depth, alpha, pos.GamePly)
			StorePvMove(currentMove, pos.GamePly)

			// store history moves
			if !pos.Capture(currentMove) {
				mFrom := FromSQ(currentMove)
				mTo := ToSQ(currentMove)
				pos.History[pos.SideToMove][mFrom][mTo] += int32(depth)
			}
			if score >= beta {
				// store hash entry with the score equal to beta
				writeHashEntry(pos.St.Top().key, int16(beta), bestMove, depth, uint8(pos.GamePly), TT_BETA)

				// store killer moves
				if !pos.Capture(currentMove) {
					pos.Killers[pos.GamePly][1] = pos.Killers[pos.GamePly][0]
					pos.Killers[pos.GamePly][0] = currentMove
				}
				return beta
			}
		}
	}

	// checkmate or stalemate is a win
	if legalMoves == 0 {
		return -int32(MATE_VALUE) + int32(pos.GamePly)
	}

	// store hash entry with the score equal to alpha
	writeHashEntry(pos.St.Top().key, int16(alpha), bestMove, depth, uint8(pos.GamePly), hashFlag)

	return alpha
}

// search position for the best move
func (pos *PositionNG) SearchPosition(depth uint8) (bestMove MoveNG) {
	clearSearch(pos)
	now := time.Now()
	var prevScore Value

	// iterative deepening
	for currentDepth := uint8(1); currentDepth <= depth; currentDepth++ {
		alpha := -VALUE_INFINITE
		beta := VALUE_INFINITE
		if currentDepth > 2 {
			window := Value(40 + 15*int(currentDepth))
			alpha = max(prevScore-window, -VALUE_INFINITE)
			beta = min(prevScore+window, VALUE_INFINITE)
		}
		score := Negamax(alpha, beta, pos, currentDepth, true)
		if currentDepth > 2 && (score <= alpha || score >= beta) {
			score = Negamax(-VALUE_INFINITE, VALUE_INFINITE, pos, currentDepth, true)
		}
		prevScore = score

		fmt.Printf("info score cp %d depth %d nodes %d time %v pv",
			score, currentDepth, pos.Nodes, time.Since(now))
		for cnt := 0; cnt < PvLength[0]; cnt++ {
			fmt.Printf(" %s", Move2Str(PvTable[cnt]))
		}
		fmt.Println()
	}
	bestMove = PvTable[0]
	return bestMove
}

func clearSearch(pos *PositionNG) {
	pos.GamePly = 0
	pos.Nodes = 0
	clear(PvTable[:])
	clear(PvLength[:])
	clear(pos.Killers[:])
	clear(pos.History[:])
}

const (
	INFINITY   int16 = 32002
	MATE_VALUE int16 = 32000
	MATE_SCORE int16 = 31000
)

const NO_HASH int16 = 32767

func readHashEntry(key Key, alpha, beta int16, bestMove *MoveNG, depth, ply uint8) (int16, MoveNG) {
	entry := &TT.Entries[key&TT.Mask]
	if entry.Key == key {
		if entry.Depth >= depth {
			score := entry.Score
			if score < -MATE_SCORE {
				score += int16(ply)
			}
			if score > MATE_SCORE {
				score -= int16(ply)
			}
			if entry.Flag == TT_EXACT {
				return score, entry.Move
			}
			if entry.Flag == TT_ALPHA && score <= alpha {
				return alpha, entry.Move
			}
			if entry.Flag == TT_BETA && score >= beta {
				return beta, entry.Move
			}
		}
	}
	return NO_HASH, MOVE_NONE
}

// write hash entry data
func writeHashEntry(key Key, score int16, bestMove MoveNG, depth, ply uint8, flag int8) {
	entry := &TT.Entries[key&TT.Mask]
	if score < -MATE_SCORE {
		score -= int16(ply)
	}
	if score > MATE_SCORE {
		score += int16(ply)
	}
	entry.Key = key
	entry.Score = score
	entry.Flag = flag
	entry.Depth = depth
	entry.Move = bestMove
	entry.Age = age
}
