# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

godogpaw is a UCCI (Universal Chinese Chess Interface) engine written in Go. It implements a full Chinese chess (Xiangqi) engine with bitboard-based position representation, negamax search with alpha-beta pruning, and the UCCI protocol for integration with chess GUIs. It also supports browser-based human-vs-AI play via Go WebAssembly.

## Build & Run Commands

```bash
# Build native binary
make native
# or: go build -o godogpaw .

# Run (listens for UCCI commands on stdin)
./godogpaw

# Build WebAssembly for browser UI
make wasm

# Serve browser UI (default port 8088)
make serve
# or: make serve PORT=9000

# Clean build artifacts
make clean

# Run all tests
go test ./engine/...

# Run a specific test
go test ./engine/ -run TestName -v

# Run benchmarks
go test ./engine/ -bench=. -benchmem
```

## Architecture

### Package Structure

- **`main.go`** — Entry point. Initializes logging (to stderr + `/tmp/godogpaw-ucci.log`), sets up panic recovery, and starts the UCCI command loop.
- **`engine/`** — Core chess engine: board representation, move generation, search, evaluation, and transposition table.
- **`ucci/`** — UCCI protocol handler. Parses stdin commands (`ucci`, `isready`, `position`, `go`, `perft`, etc.) and dispatches to the engine.
- **`wasm/`** — WebAssembly entry point (`wasm/main.go`). Exposes engine functions to JavaScript: `engineNewGame`, `engineGetBoard`, `engineGetLegalMovesFrom`, `engineDoMoveBySquares`, `engineUndoMove`, `engineSearch`.
- **`web/`** — Browser UI files: `index.html`, `app.js`, `board.js`, `style.css`. Communicates with the engine via the WASM bridge.

### Engine Internals

**Board Representation (`bitboard.go`, `positionng.go`):**
The 90-square Xiangqi board uses a custom 128-bit bitboard (two `uint64` fields: `Lo`, `Hi`). `PositionNG` is the central struct holding board state, piece bitboards by type/color, king squares, state history stack, and search-related tables (killers, history, bloom filter).

**Move Encoding (`types.go`):**
Moves are 16-bit: 7 bits source + 7 bits destination. Pieces use color+type encoding (e.g., `W_ROOK`, `B_CANNON`). `HistoryTable` is `[COLOR_NB][SQUARE_NB][SQUARE_NB]Value`.

**Magic Bitboards (`magic.go`):**
Pre-computed magic number tables for fast rook/cannon sliding attack generation, adapted for the 9×10 board.

**Move Generation (`positionng.go`):**
`Generate(GenType, moveList)` produces pseudo-legal moves by type (CAPTURES, QUIETS, EVASIONS, etc.). `GenerateLEGAL` filters for legality. Chinese chess-specific rules: cannon jumping captures, knight leg blocking, restricted bishop/advisor movement, pawn directional attacks changing after crossing the river.

**Move Ordering (`movepicker.go`):**
`MovePicker` (inspired by Ethereal) uses staged move generation: TT move → noisy moves (MVV-LVA scored) → killer 1 → killer 2 → quiet moves (history-scored). Quiescence search skips quiet stages.

**Search (`search.go`):**
Iterative deepening negamax with aspiration windows (dynamic window: `40 + 15*depth`). Features:
- Null move pruning (depth > 2, static eval >= beta)
- Mate distance pruning
- Razoring (depth 1 and depth < 4)
- Futility pruning (depth < 4, non-mate scores)
- Late Move Reductions (LMR) at depth >= 5, for non-killer quiet moves after 3+ moves searched
- Principal Variation Search (PVS) at depth >= 5
- Check extension (+1 depth when in check)
- Quiescence search for captures only
- Repetition detection with penalty (`-MATERIAL_WEIGHTS[W_CANNON]`)

**UCI Move Parsing (`uci_helpers.go`):**
`SquareFromString` and `ParseUCIMove` convert coordinate notation (e.g., "b2e2") to internal `MoveNG`, validating against legal moves.

**History Table (`history.go`):**
`GetHistoryScore` retrieves history heuristic scores indexed by `[color][from][to]`.

**Transposition Table (`tt.go`, `tt_js.go`):**
16 MB default on native; 2 MB on WASM builds (via `tt_js.go` init override). Entries store hash key, score, depth, bound flag (TT_ALPHA/TT_BETA/TT_EXACT), best move, and age for replacement policy.

**Evaluation (`evaluation.go`):**
Piece-square tables per piece type on the 9×10 board, plus material weights (`MATERIAL_WEIGHTS` array). Black squares are flipped via `flipSquare`. Score relative to side-to-move with a small `advancedValue` tempo bonus.

**WASM Yield (`yield_js.go`, `yield_other.go`):**
On WASM builds, `MaybeYield()` calls `runtime.Gosched()` every ~8192 nodes to prevent browser UI freezes during search. No-op on native builds.

### Key Constants

- `MAX_MOVES = 128` (max moves per position)
- `MAX_PLY = 246` (max search depth)
- `VALUE_MATE = 32000`
- `MATE_VALUE = 32000`, `MATE_SCORE = 31000` (mate detection thresholds in search)
- Starting FEN: `rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1`

### Incomplete Features

- UCCI commands `setoption`, `banmoves`, `ponderhit`, `stop` are stubs (TODO).
- `go` command only supports `go depth N`; no time management.
