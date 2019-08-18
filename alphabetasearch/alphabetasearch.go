package alphabetasearch

type Move = int32

type Board interface {
	IsMaximizingPlayerTurn() bool
	AllMoves() []Move
	AllMovesCheckLegal() []Move
	MakeMove(Move)
	UnMakeMove(Move)
	Evaluate() int
	ProbeHash(depth uint8) (bestMove Move, score int, bound int8, ok bool)
	RecordHash(depth uint8, score int16, move Move, bound int8)
}

const (
	HashAlpha int8 = iota
	HashBeta
	HashPv
)

func AlphaBetaSearch(board Board, depth uint8, alpha, beta int) (bestMove Move, score int) {
	bestMoveHash, scoreHash, bound, ok := board.ProbeHash(depth)
	if ok {
		switch bound {
		case HashBeta:
			if scoreHash >= beta {
				return bestMoveHash, scoreHash
			}
		case HashAlpha:
			if scoreHash <= alpha {
				return bestMoveHash, scoreHash
			}
		case HashPv:
			return bestMoveHash, scoreHash
		}
	}

	var hashFlag int8 = HashAlpha
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
			board.RecordHash(depth, int16(beta), moves[i], HashBeta)
			return moves[i], beta
		}
		if value > alpha {
			alpha = value
			bestMove = moves[i]
			hashFlag = HashPv
		}
	}
	board.RecordHash(depth, int16(alpha), bestMove, hashFlag)
	return bestMove, alpha
}

func alphaBetaSearch(board Board, depth uint8, alpha, beta int) (score int) {
	_, scoreHash, bound, ok := board.ProbeHash(depth)
	if ok {
		switch bound {
		case HashBeta:
			if scoreHash >= beta {
				return scoreHash
			}
		case HashAlpha:
			if scoreHash <= alpha {
				return scoreHash
			}
		case HashPv:
			return scoreHash
		}
	}

	var hashFlag int8 = HashAlpha
	if depth == 0 {
		score = board.Evaluate()
		// board.RecordHash(depth, int16(score), 0)
		return score
	}
	var bestMove Move
	moves := board.AllMoves()
	for _, move := range moves {
		board.MakeMove(move)
		value := -alphaBetaSearch(board, depth-1, -beta, -alpha)
		board.UnMakeMove(move)
		if value >= beta {
			board.RecordHash(depth, int16(beta), move, HashBeta)
			return beta
		}
		if value > alpha {
			alpha = value
			bestMove = move
		}
	}
	board.RecordHash(depth, int16(alpha), bestMove, hashFlag)
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
	_, scoreHash, bound, ok := board.ProbeHash(depth)
	if ok {
		switch bound {
		case HashBeta:
			if scoreHash >= beta {
				return scoreHash
			}
		case HashAlpha:
			if scoreHash <= alpha {
				return scoreHash
			}
		case HashPv:
			return scoreHash
		}
	}

	var hashFlag int8 = HashAlpha
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
	var bestMove Move
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
			bestMove = move
			alpha = value
			hashFlag = HashPv
		}
		if alpha >= beta {
			hashFlag = HashBeta
			break
		}
	}
	board.RecordHash(depth, int16(alpha), bestMove, hashFlag)
	return alpha
}
