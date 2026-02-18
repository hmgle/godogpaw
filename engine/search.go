package engine

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

var (
	PvTable  = [MAX_MOVES * MAX_MOVES]MoveNG{}
	PvLength = [MAX_MOVES]int{}
)

// Pre-computed LMR reduction table: lmrTable[depth][moveCount].
var lmrTable [64][64]int

func init() {
	for d := 1; d < 64; d++ {
		for m := 1; m < 64; m++ {
			lmrTable[d][m] = int(0.75 + math.Log(float64(d))*math.Log(float64(m))/2.25)
		}
	}
}

// SearchLimits holds the time and depth constraints for a search.
type SearchLimits struct {
	Depth     uint8
	TimeLimit time.Duration // 0 means no time limit
	Infinite  bool
}

// stopFlag is set to 1 when the search should be aborted.
var stopFlag atomic.Int32

// StopSearch sets the flag to abort the search.
func StopSearch() {
	stopFlag.Store(1)
}

func shouldStop() bool {
	return stopFlag.Load() != 0
}

// Repetition penalty value (approximately cannon value).
const repetitionPenalty Value = 640

func StorePvMove(move MoveNG, searchPly int) {
	PvTable[searchPly*int(MAX_MOVES)+searchPly] = move
	for nextPly := searchPly + 1; nextPly < PvLength[searchPly+1]; nextPly++ {
		PvTable[searchPly*int(MAX_MOVES)+nextPly] = PvTable[(searchPly+1)*int(MAX_MOVES)+nextPly]
	}
	PvLength[searchPly] = PvLength[searchPly+1]
}

func Quiescence(alpha, beta Value, pos *PositionNG) (bestScore Value) {
	PvLength[pos.GamePly] = pos.GamePly
	evaluation := pos.Evaluate()
	if pos.GamePly >= int(MAX_MOVES) {
		return evaluation
	}
	if evaluation >= beta {
		return evaluation
	}
	if evaluation > alpha {
		alpha = evaluation
	}

	var mp MovePicker
	InitalizeMovePicker(&mp, true, MOVE_NONE, MOVE_NONE, MOVE_NONE, MOVE_NONE, &pos.History)
	for currentMove := SelectNextMove(&mp, pos); currentMove != MOVE_NONE; currentMove = SelectNextMove(&mp, pos) {
		if !pos.Legal(currentMove) {
			continue
		}
		// Skip losing captures in quiescence (SEE < 0)
		if pos.SEE(currentMove) < 0 {
			continue
		}
		var st StateInfo
		pos.DoMove(currentMove, &st)
		score := -Quiescence(-beta, -alpha, pos)
		pos.UndoMove(currentMove)
		if score > alpha {
			StorePvMove(currentMove, pos.GamePly)
			alpha = score

			if score >= beta {
				return score
			}
		}
	}
	return alpha
}

func Negamax(alpha, beta Value, pos *PositionNG, depth uint8, doNullMove bool) (bestScore Value) {
	MaybeYield()
	PvLength[pos.GamePly] = pos.GamePly
	rootNode := pos.GamePly == 0
	pvNode := alpha != beta-1
	hashFlag := TT_ALPHA
	var score Value
	var legalMoves int

	if pos.IsDraw() {
		return 0
	}

	// Check time periodically (every 4096 nodes at root level)
	if pos.Nodes&4095 == 0 && shouldStop() {
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
		return -repetitionPenalty
	}
	if depth == 0 {
		return Quiescence(alpha, beta, pos)
	}

	// Mate distance pruning
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
	if !pvNode && !inCheck {
		staticEval = pos.Evaluate()

		// Razoring: if static eval + margin < alpha and depth is low, drop into QSearch
		if depth <= 3 && !rootNode {
			razoringMargin := Value(300 + 200*int(depth))
			if staticEval+razoringMargin < alpha {
				if depth == 1 {
					return Quiescence(alpha, beta, pos)
				}
				qScore := Quiescence(alpha, beta, pos)
				if qScore < alpha {
					return qScore
				}
			}
		}

		// Reverse futility pruning (static eval too high → prune)
		if depth <= 5 && !rootNode && staticEval-Value(80*int(depth)) >= beta {
			return staticEval
		}

		// Null move pruning with adaptive R
		if doNullMove && pos.GamePly > 0 && depth > 2 && staticEval >= beta {
			r := uint8(3 + depth/6)
			if r > depth-1 {
				r = depth - 1
			}
			var st StateInfo
			pos.DoNullMove(&st)
			score = -Negamax(-beta, -beta+1, pos, depth-1-r, false)
			pos.UndoNullMove()

			if score >= beta {
				return beta
			}
		}
	}

	// Futility pruning condition
	futilityPruning := false
	if !pvNode && !inCheck && depth <= 3 {
		futilityMargin := Value(200 * int(depth))
		staticEvalForFutility := staticEval
		if staticEvalForFutility == 0 && (pvNode || inCheck) {
			staticEvalForFutility = pos.Evaluate()
		}
		if abs(int(alpha)) < int(MATE_SCORE) && staticEvalForFutility+futilityMargin <= alpha {
			futilityPruning = true
		}
	}

	// Determine counter-move
	var counterMove MoveNG
	if pos.GamePly > 0 {
		prevSt := pos.St.Prev()
		if prevSt.capturedPiece == NO_PIECE {
			// Find the piece that moved to the previous destination
			// We can derive this from the current board state
			// The previous move's "to" square now contains the piece that was moved
			// But we need to know the previous move's to-square
			// Since we don't have direct access, use counter-move table if available
		}
		counterMove = pos.GetCounterMove()
	}

	movesSearched := 0
	var quietsSearched [64]MoveNG
	quietCount := 0

	// Loop over moves
	var mp MovePicker
	InitalizeMovePicker(&mp, false, ttMove, pos.Killers[pos.GamePly][0], pos.Killers[pos.GamePly][1], counterMove, &pos.History)
	for currentMove := SelectNextMove(&mp, pos); currentMove != MOVE_NONE; currentMove = SelectNextMove(&mp, pos) {
		if !pos.Legal(currentMove) {
			continue
		}
		legalMoves++

		isCapture := pos.Capture(currentMove)
		givesCheck := pos.GivesCheck(currentMove)

		// Futility pruning: skip quiet non-checking moves when eval is too low
		if futilityPruning && movesSearched > 0 && !isCapture && !givesCheck {
			continue
		}

		var st StateInfo
		pos.DoMove(currentMove, &st)

		// PVS + LMR
		if movesSearched == 0 {
			// First move: full window search
			score = -Negamax(-beta, -alpha, pos, depth-1, true)
		} else {
			reduction := uint8(0)

			// LMR: reduce depth for late quiet moves
			if depth >= 3 && movesSearched >= 3 && !isCapture && !inCheck {
				d := min(int(depth), 63)
				m := min(movesSearched, 63)
				reduction = uint8(lmrTable[d][m])

				// Reduce less for killer moves
				if currentMove == pos.Killers[pos.GamePly][0] || currentMove == pos.Killers[pos.GamePly][1] {
					if reduction > 0 {
						reduction--
					}
				}

				// Adjust by history score
				histScore := GetHistoryScore(&pos.History, currentMove, pos.SideToMove)
				if histScore > 0 && reduction > 0 {
					reduction--
				} else if histScore < -100 {
					reduction++
				}

				// Don't reduce below 1
				if reduction >= depth-1 {
					reduction = depth - 2
				}
			}

			// PVS: search with null window
			score = -Negamax(-alpha-1, -alpha, pos, depth-1-reduction, true)

			// Re-search at full depth if LMR failed high
			if reduction > 0 && score > alpha {
				score = -Negamax(-alpha-1, -alpha, pos, depth-1, true)
			}

			// Re-search with full window if PVS failed high in PV nodes
			if score > alpha && score < beta {
				score = -Negamax(-beta, -alpha, pos, depth-1, true)
			}
		}

		pos.UndoMove(currentMove)
		movesSearched++

		// Check for search abort
		if shouldStop() {
			return 0
		}

		if score > alpha {
			hashFlag = TT_EXACT
			bestMove = currentMove
			alpha = score
			StorePvMove(currentMove, pos.GamePly)

			if score >= beta {
				// Store hash entry with beta flag
				writeHashEntry(pos.St.Top().key, int16(beta), bestMove, depth, uint8(pos.GamePly), TT_BETA)

				if !isCapture {
					// Store killer moves
					pos.Killers[pos.GamePly][1] = pos.Killers[pos.GamePly][0]
					pos.Killers[pos.GamePly][0] = currentMove

					// Update history with depth^2 bonus
					bonus := int32(depth) * int32(depth)
					UpdateHistory(&pos.History, currentMove, pos.SideToMove, bonus)

					// History malus: penalize previously searched quiet moves
					for i := 0; i < quietCount; i++ {
						UpdateHistory(&pos.History, quietsSearched[i], pos.SideToMove, -bonus)
					}

					// Store counter-move
					pos.StoreCounterMove(currentMove)
				}
				return beta
			}
		}

		// Track searched quiet moves for history malus
		if !isCapture && quietCount < 64 {
			quietsSearched[quietCount] = currentMove
			quietCount++
		}
	}

	// Checkmate or stalemate is a win (for the opponent in xiangqi)
	if legalMoves == 0 {
		return -int32(MATE_VALUE) + int32(pos.GamePly)
	}

	// Store hash entry with the score
	writeHashEntry(pos.St.Top().key, int16(alpha), bestMove, depth, uint8(pos.GamePly), hashFlag)

	return alpha
}

// SearchPosition searches for the best move with the given limits.
func (pos *PositionNG) SearchPosition(depth uint8) (bestMove MoveNG) {
	return pos.SearchPositionWithLimits(SearchLimits{Depth: depth})
}

// SearchPositionWithLimits searches with time and depth constraints.
func (pos *PositionNG) SearchPositionWithLimits(limits SearchLimits) (bestMove MoveNG) {
	clearSearch(pos)
	stopFlag.Store(0)
	now := time.Now()
	var prevScore Value

	// Set up time-based stop
	if limits.TimeLimit > 0 {
		go func() {
			time.Sleep(limits.TimeLimit)
			stopFlag.Store(1)
		}()
	}

	maxDepth := limits.Depth
	if maxDepth == 0 {
		maxDepth = uint8(MAX_PLY)
	}

	// Iterative deepening
	for currentDepth := uint8(1); currentDepth <= maxDepth; currentDepth++ {
		alpha := -VALUE_INFINITE
		beta := VALUE_INFINITE

		// Aspiration windows for depth > 2
		if currentDepth > 2 {
			window := Value(30)
			alpha = max(prevScore-window, -VALUE_INFINITE)
			beta = min(prevScore+window, VALUE_INFINITE)
		}

		score := Negamax(alpha, beta, pos, currentDepth, true)

		// Re-search with wider windows if aspiration failed
		if currentDepth > 2 && (score <= alpha || score >= beta) {
			// Widen window by 2x
			window := Value(60)
			alpha = max(prevScore-window, -VALUE_INFINITE)
			beta = min(prevScore+window, VALUE_INFINITE)
			score = Negamax(alpha, beta, pos, currentDepth, true)

			// Full window if still failing
			if score <= alpha || score >= beta {
				score = Negamax(-VALUE_INFINITE, VALUE_INFINITE, pos, currentDepth, true)
			}
		}

		// If search was stopped mid-iteration, use best move from last complete iteration
		if shouldStop() && currentDepth > 1 {
			break
		}

		prevScore = score
		bestMove = PvTable[0]

		fmt.Printf("info score cp %d depth %d nodes %d time %v pv",
			score, currentDepth, pos.Nodes, time.Since(now))
		for cnt := 0; cnt < PvLength[0]; cnt++ {
			fmt.Printf(" %s", Move2Str(PvTable[cnt]))
		}
		fmt.Println()
	}

	if bestMove == MOVE_NONE {
		bestMove = PvTable[0]
	}
	return bestMove
}

func clearSearch(pos *PositionNG) {
	pos.GamePly = 0
	pos.Nodes = 0
	clear(PvTable[:])
	clear(PvLength[:])
	clear(pos.Killers[:])
	// Don't clear history between searches - it accumulates useful data
	// But apply aging (divide by 2)
	for c := 0; c < COLOR_NB; c++ {
		for f := 0; f < SQUARE_NB; f++ {
			for t := 0; t < SQUARE_NB; t++ {
				pos.History[c][f][t] /= 2
			}
		}
	}
	// Increment TT age
	age++
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
		*bestMove = entry.Move
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
		return NO_HASH, entry.Move
	}
	return NO_HASH, MOVE_NONE
}

// writeHashEntry stores data in the transposition table with age-based replacement.
func writeHashEntry(key Key, score int16, bestMove MoveNG, depth, ply uint8, flag int8) {
	entry := &TT.Entries[key&TT.Mask]
	if score < -MATE_SCORE {
		score -= int16(ply)
	}
	if score > MATE_SCORE {
		score += int16(ply)
	}

	// Replace if: different age (old search), higher depth, or exact bound
	if entry.Age != age || entry.Depth <= depth || flag == TT_EXACT {
		entry.Key = key
		entry.Score = score
		entry.Flag = flag
		entry.Depth = depth
		entry.Move = bestMove
		entry.Age = age
	}
}
