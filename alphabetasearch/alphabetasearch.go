package alphabetasearch

type Move = int32

type Board interface {
	IsMaximizingPlayerTurn() bool
	AllMoves() []Move
	AllMovesCheckLegal() []Move
	MakeMove(Move)
	UnMakeMove(Move)
	Evaluate() int
	ProbeHash(depth uint8) (bestMove Move, score int, ok bool)
	RecordHash(depth uint8, score int16, move Move)
}

func AlphaBetaSearch(board Board, depth uint8, alpha, beta int) (bestMove Move, score int) {
	/*
		bestMoveHash, scoreHash, ok := board.ProbeHash(depth)
		if ok {
			return bestMoveHash, scoreHash
		}
	*/

	if depth == 0 {
		score = board.Evaluate()
		// board.RecordHash(depth, int16(score), 0)
		return 0, score
	}

	moves := board.AllMovesCheckLegal()
	for i, move := range moves {
		board.MakeMove(move)
		// value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
		// value := -negaScoutSearch(board, depth-1, -beta, -alpha)
		value := -pvsSearch(board, depth-1, -beta, -alpha)
		board.UnMakeMove(move)

		if value >= beta {
			// board.RecordHash(depth, int16(beta), moves[i])
			return moves[i], beta
		}
		if value > alpha {
			alpha = value
			bestMove = moves[i]
		}
	}
	// board.RecordHash(depth, int16(alpha), bestMove)
	return bestMove, alpha
}

func alphaBetaSearch(board Board, depth uint8, alpha, beta int) (score int) {
	/*
		_, scoreHash, ok := board.ProbeHash(depth)
		if ok {
			return scoreHash
		}
	*/

	if depth == 0 {
		score = board.Evaluate()
		// board.RecordHash(depth, int16(score), 0)
		return score
	}
	// var bestMove Move
	moves := board.AllMoves()
	for _, move := range moves {
		board.MakeMove(move)
		value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
		board.UnMakeMove(move)
		if value >= beta {
			// board.RecordHash(depth, int16(beta), move)
			return beta
		}
		if value > alpha {
			alpha = value
			// bestMove = move
		}
	}
	// board.RecordHash(depth, int16(alpha), bestMove)
	return alpha
}

func negaScoutSearch(board Board, depth uint8, alpha, beta int) (score int) {
	if depth == 0 {
		return board.Evaluate()
	}
	moves := board.AllMoves()
	if len(moves) == 0 {
		return alpha
	}
	m := moves[0]
	board.MakeMove(m)
	current := -negaScoutSearch(board, depth-1, -beta, -alpha)
	board.UnMakeMove(m)
	for _, move := range moves[1:] {
		board.MakeMove(move)
		value := -negaScoutSearch(board, depth-1, -alpha-1, -alpha)
		if value > alpha && value < beta {
			value = -negaScoutSearch(board, depth-1, -beta, -alpha)
		}
		board.UnMakeMove(move)
		if value >= current {
			current = value
			if value >= alpha {
				alpha = value
			}
			if value >= beta {
				break
			}
		}
	}
	return current
}

func pvsSearch(board Board, depth uint8, alpha, beta int) (score int) {
	/*
		_, scoreHash, ok := board.ProbeHash(depth)
		if ok {
			return scoreHash
		}
	*/
	if depth == 0 {
		score = board.Evaluate()
		// board.RecordHash(depth, int16(score), 0)
		return score
	}
	moves := board.AllMoves()
	if len(moves) == 0 {
		return alpha
	}
	var value int
	// var bestMove Move
	for i, move := range moves {
		board.MakeMove(move)
		if i == 0 {
			value = -pvsSearch(board, depth-1, -beta, -alpha)
		} else {
			value = -pvsSearch(board, depth-1, -alpha-1, -alpha)
			if value > alpha && value < beta {
				value = -pvsSearch(board, depth-1, -beta, -value)
			}
		}
		board.UnMakeMove(move)
		if value > alpha {
			// bestMove = move
			alpha = value
		}
		if alpha >= beta {
			break
		}
	}
	// board.RecordHash(depth, int16(alpha), bestMove)
	return alpha
}
