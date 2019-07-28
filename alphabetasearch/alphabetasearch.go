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
		value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
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
