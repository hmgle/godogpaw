package alphabetasearch

type Move = int32

type Board interface {
	IsMaximizingPlayerTurn() bool
	AllMoves() []Move
	AllMovesCheckLegal() []Move
	MakeMove(Move)
	UnMakeMove(Move)
	Evaluate() int
}

func AlphaBetaSearch(board Board, depth uint8, alpha, beta int) (bestMove Move, score int) {
	if depth == 0 {
		return 0, board.Evaluate()
	}

	moves := board.AllMovesCheckLegal()
	for i, move := range moves {
		board.MakeMove(move)
		// value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
		// value := -negaScoutSearch(board, depth-1, -beta, -alpha)
		value := -pvsSearch(board, depth-1, -beta, -alpha)
		board.UnMakeMove(move)

		if value >= beta {
			return moves[i], beta
		}
		if value > alpha {
			alpha = value
			bestMove = moves[i]
		}
	}
	return bestMove, alpha
}

func alphaBetaSearch(board Board, depth uint8, alpha, beta int) (score int) {
	if depth == 0 {
		return board.Evaluate()
	}
	moves := board.AllMoves()
	for _, move := range moves {
		board.MakeMove(move)
		value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
		board.UnMakeMove(move)
		if value >= beta {
			return beta
		}
		if value > alpha {
			alpha = value
		}
	}
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
	if depth == 0 {
		return board.Evaluate()
	}
	moves := board.AllMoves()
	if len(moves) == 0 {
		return alpha
	}
	var value int
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
			alpha = value
		}
		if alpha >= beta {
			break
		}
	}
	return alpha
}
