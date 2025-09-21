package engine

import (
	"fmt"
	"strings"
)

// SquareFromString converts a coordinate like "a0" to the internal Square type.
func SquareFromString(coord string) (Square, error) {
	if len(coord) != 2 {
		return SQ_NONE, fmt.Errorf("invalid square: %q", coord)
	}
	file := coord[0]
	rank := coord[1]
	if file < 'a' || file > 'i' {
		return SQ_NONE, fmt.Errorf("invalid file: %q", coord)
	}
	if rank < '0' || rank > '9' {
		return SQ_NONE, fmt.Errorf("invalid rank: %q", coord)
	}
	f := File(file - 'a')
	r := Rank(rank - '0')
	if f < FILE_A || f >= FILE_NB || r < RANK_0 || r >= RANK_NB {
		return SQ_NONE, fmt.Errorf("square out of range: %q", coord)
	}
	return MakeSquareNG(f, r), nil
}

// ParseUCIMove converts a coordinate move string (e.g. "b2e2") into a legal move for the given position.
func ParseUCIMove(pos *PositionNG, moveStr string) (MoveNG, error) {
	moveStr = strings.TrimSpace(strings.ToLower(moveStr))
	if len(moveStr) < 4 {
		return MOVE_NONE, fmt.Errorf("move too short: %q", moveStr)
	}
	from, err := SquareFromString(moveStr[:2])
	if err != nil {
		return MOVE_NONE, err
	}
	to, err := SquareFromString(moveStr[2:4])
	if err != nil {
		return MOVE_NONE, err
	}
	var list [MAX_MOVES]MoveNG
	size := pos.GenerateLEGAL(list[:])
	for i := uint8(0); i < size; i++ {
		mv := list[i]
		if FromSQ(mv) == from && ToSQ(mv) == to {
			return mv, nil
		}
	}
	return MOVE_NONE, fmt.Errorf("illegal move %q for current position", moveStr)
}
