//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/hmgle/godogpaw/engine"
)

const startFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

var pos engine.PositionNG

type moveRecord struct {
	move engine.MoveNG
	st   engine.StateInfo
}

var moveHistory []moveRecord

// boardState is the JSON-serializable snapshot returned by engineGetBoard.
type boardState struct {
	Board      [engine.SQUARE_NB]int `json:"board"`
	SideToMove int                   `json:"sideToMove"`
	InCheck    bool                  `json:"inCheck"`
	IsGameOver bool                  `json:"isGameOver"`
	LastFrom   int                   `json:"lastMoveFrom"`
	LastTo     int                   `json:"lastMoveTo"`
}

func engineNewGame(_ js.Value, args []js.Value) any {
	fen := startFEN
	if len(args) > 0 {
		s := args[0].String()
		if s != "" {
			fen = s
		}
	}
	pos.Set(fen)
	moveHistory = nil
	return nil
}

func engineGetBoard(_ js.Value, _ []js.Value) any {
	var st boardState
	for i := 0; i < engine.SQUARE_NB; i++ {
		st.Board[i] = pos.Board[i]
	}
	st.SideToMove = int(pos.SideToMove)
	st.InCheck = pos.Checkers().IsNotZero()

	// Check game over: no legal moves
	var list [engine.MAX_MOVES]engine.MoveNG
	size := pos.GenerateLEGAL(list[:])
	st.IsGameOver = size == 0

	st.LastFrom = -1
	st.LastTo = -1
	if len(moveHistory) > 0 {
		last := moveHistory[len(moveHistory)-1]
		st.LastFrom = engine.FromSQ(last.move)
		st.LastTo = engine.ToSQ(last.move)
	}

	b, _ := json.Marshal(st)
	return string(b)
}

func engineGetLegalMovesFrom(_ js.Value, args []js.Value) any {
	sq := args[0].Int()
	if sq < 0 || sq >= engine.SQUARE_NB {
		return "[]"
	}

	var list [engine.MAX_MOVES]engine.MoveNG
	size := pos.GenerateLEGAL(list[:])

	targets := make([]int, 0, 8)
	for i := uint8(0); i < size; i++ {
		if engine.FromSQ(list[i]) == sq {
			targets = append(targets, engine.ToSQ(list[i]))
		}
	}
	b, _ := json.Marshal(targets)
	return string(b)
}

func engineDoMoveBySquares(_ js.Value, args []js.Value) any {
	from := args[0].Int()
	to := args[1].Int()
	if from < 0 || from >= engine.SQUARE_NB || to < 0 || to >= engine.SQUARE_NB {
		return false
	}

	m := engine.MakeMove(from, to)

	// Validate the move is legal
	var list [engine.MAX_MOVES]engine.MoveNG
	size := pos.GenerateLEGAL(list[:])
	legal := false
	for i := uint8(0); i < size; i++ {
		if list[i] == m {
			legal = true
			break
		}
	}
	if !legal {
		return false
	}

	var rec moveRecord
	rec.move = m
	pos.DoMove(m, &rec.st)
	moveHistory = append(moveHistory, rec)
	return true
}

func engineUndoMove(_ js.Value, _ []js.Value) any {
	if len(moveHistory) == 0 {
		return false
	}
	last := moveHistory[len(moveHistory)-1]
	pos.UndoMove(last.move)
	moveHistory = moveHistory[:len(moveHistory)-1]
	return true
}

func engineSearch(_ js.Value, args []js.Value) any {
	depth := uint8(4)
	if len(args) > 0 && args[0].Int() > 0 {
		depth = uint8(args[0].Int())
	}

	// Return a Promise so JS can await the result
	handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		go func() {
			bestMove := pos.SearchPosition(depth)
			if !engine.IsOKMove(bestMove) {
				resolve.Invoke("")
				return
			}

			// Execute the best move
			var rec moveRecord
			rec.move = bestMove
			pos.DoMove(bestMove, &rec.st)
			moveHistory = append(moveHistory, rec)

			moveStr := engine.Move2Str(bestMove)
			resolve.Invoke(moveStr)
		}()
		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func main() {
	g := js.Global()
	g.Set("engineNewGame", js.FuncOf(engineNewGame))
	g.Set("engineGetBoard", js.FuncOf(engineGetBoard))
	g.Set("engineGetLegalMovesFrom", js.FuncOf(engineGetLegalMovesFrom))
	g.Set("engineDoMoveBySquares", js.FuncOf(engineDoMoveBySquares))
	g.Set("engineUndoMove", js.FuncOf(engineUndoMove))
	g.Set("engineSearch", js.FuncOf(engineSearch))

	// Initialize with default starting position
	pos.Set(startFEN)

	// Keep the Go program running
	select {}
}
