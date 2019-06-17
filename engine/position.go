package engine

import (
	"fmt"

	"github.com/willf/bitset"
)

// Position 位置信息.
type Position struct {
	Pawns    *bitset.BitSet
	Cannons  *bitset.BitSet
	Rooks    *bitset.BitSet
	Knights  *bitset.BitSet
	Bishops  *bitset.BitSet
	Advisors *bitset.BitSet
	Kings    *bitset.BitSet

	Red   *bitset.BitSet
	Black *bitset.BitSet

	Checkers *bitset.BitSet

	IsRedMove bool
	// Key 当前局面哈希
	Key uint64
}

// WhatPiece 返回 sq 位置的棋子类型.
func (p *Position) WhatPiece(sq uint) int {
	if !p.Red.Test(sq) && !p.Black.Test(sq) {
		return Empty
	}
	if p.Pawns.Test(sq) {
		return Pawn
	}
	if p.Rooks.Test(sq) {
		return Rook
	}
	if p.Cannons.Test(sq) {
		return Cannon
	}
	if p.Knights.Test(sq) {
		return Knight
	}
	if p.Bishops.Test(sq) {
		return Bishop
	}
	if p.Advisors.Test(sq) {
		return Advisor
	}
	if p.Kings.Test(sq) {
		return King
	}
	panic(fmt.Errorf("Wrong piece on %d", sq))
}
