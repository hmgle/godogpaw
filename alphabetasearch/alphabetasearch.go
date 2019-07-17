package alphabetasearch

import "github.com/sirupsen/logrus"

type Move = int32

type Board interface {
	IsMaximizingPlayerTurn() bool
	AllMoves() []Move
	MakeMove(Move)
	UnMakeMove(Move)
	Evaluate() int
}

func AlphaBetaSearch(board Board, depth uint8, alpha, beta int) (bestMove Move, score int) {
	if depth == 0 {
		return 0, board.Evaluate()
	}
	moves := board.AllMoves()
	if board.IsMaximizingPlayerTurn() {
		for i, move := range moves {
			board.MakeMove(move)
			value := alphaBetaSearch(board, depth-1, alpha, beta)
			board.UnMakeMove(move)
			if value > alpha {
				alpha = value
				bestMove = moves[i]
			}
			if alpha >= beta {
				return moves[i], alpha
			}
		}
		return bestMove, alpha
	} else {
		for i, move := range moves {
			board.MakeMove(move)
			value := alphaBetaSearch(board, depth-1, alpha, beta)
			board.UnMakeMove(move)
			if value < beta {
				beta = value
				bestMove = moves[i]
			}
			if alpha >= beta {
				logrus.Debugf("alpha: %d, beta: %d, movi: %d\n", alpha, beta, i)
				return moves[i], alpha
			}
		}
		logrus.Debugf("alpha: %d, beta: %d\n", alpha, beta)
		return bestMove, beta
	}
}

func alphaBetaSearch(board Board, depth uint8, alpha, beta int) (score int) {
	if depth == 0 {
		return board.Evaluate()
	}
	moves := board.AllMoves()
	if len(moves) == 0 {
		return board.Evaluate()
	}
	if board.IsMaximizingPlayerTurn() {
		for _, move := range moves {
			board.MakeMove(move)
			value := alphaBetaSearch(board, depth-1, alpha, beta)
			board.UnMakeMove(move)
			if value > alpha {
				alpha = value
			}
			if beta <= alpha {
				break
			}
		}
		return alpha
	} else {
		for _, move := range moves {
			board.MakeMove(move)
			value := alphaBetaSearch(board, depth-1, alpha, beta)
			board.UnMakeMove(move)
			if value < beta {
				beta = value
			}
			if beta <= alpha {
				break
			}
		}
		return beta
	}
}
