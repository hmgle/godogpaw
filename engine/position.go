package engine

import (
	"fmt"
	"log"

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

// isKingCheck 返回将帅是否照面.
func (p *Position) isKingCheck() bool {
	// 是否同一列
	redKing, found := p.Kings.NextSet(0)
	if !found {
		log.Fatalf("not found red king")
	}
	blackKing, found := p.Kings.NextSet(redKing + 1)
	if !found {
		log.Fatalf("not found black king")
	}
	if File(int(redKing)) != File(int(blackKing)) {
		return false
	}

	return !p.IsAnyPieceBetweenFile(int(redKing), int(blackKing))

	/*
		// 中间是否有阻挡
		file := File(int(redKing))
		gb := p.Black.Union(p.Red)
		fileMask := gb.Intersection(FileMasks[file])
		if fileMask.Count() == 2 { // 除了将帅外，没有阻挡
			return true
		}
		for i, e := fileMask.NextSet(redKing); e; i, e = fileMask.NextSet(i + 1) {
			if i > redKing && i < blackKing {
				// 中间有阻挡
				return false
			}
			if i == blackKing {
				break
			}
		}
		return true
	*/
}

// IsAnyPieceBetweenFile 判断同一列的两个棋子(sq1, sq2)之间是否还有其他棋子.
// sq1, sq2 必须是同一列的，即 File(sq1) == File(sq2).
func (p *Position) IsAnyPieceBetweenFile(sq1, sq2 int) bool {
	file := File(sq1)
	if File(sq2) != file {
		log.Fatalf("sq1(%d) and sq2(%d) is not in same file", sq1, sq2)
	}

	min, max := sq1, sq2
	if sq1 > sq2 {
		min, max = sq2, sq1
	}
	gb := p.Black.Union(p.Red)
	fileMask := gb.Intersection(FileMasks[file])
	next, _ := fileMask.NextSet(uint(min + 1))
	return int(next) < max
}

// IsAnyPieceBetweenRank 判断同一行的两个棋子(sq1, sq2)之间是否还有其他棋子.
// sq1, sq2 必须是同一行的，即 Rank(sq1) == Rank(sq2).
func (p *Position) IsAnyPieceBetweenRank(sq1, sq2 int) bool {
	rank := Rank(sq1)
	if Rank(sq2) != rank {
		log.Fatalf("sq1(%d) and sq2(%d) is not in same rank", sq1, sq2)
	}

	min, max := sq1, sq2
	if sq1 > sq2 {
		min, max = sq2, sq1
	}
	gb := p.Black.Union(p.Red)
	rankMask := gb.Intersection(RankMasks[rank])
	next, _ := rankMask.NextSet(uint(min + 1))
	return int(next) < max
}

// isRookCheck 返回是否车将.
func (p *Position) isRookCheck() bool {
	var (
		kingSq uint
		selfPs *bitset.BitSet
		sidePs *bitset.BitSet
	)
	if p.IsRedMove {
		selfPs = p.Red
		sidePs = p.Black
	} else {
		selfPs = p.Black
		sidePs = p.Red
	}
	kingSq, _ = p.Kings.Intersection(selfPs).NextSet(0)
	rookAttacks := RookAttacks[int(kingSq)]
	sideRooks := p.Rooks.Intersection(sidePs)
	// 先判断是否己方帅同一行及同一列有没有对方车
	if !rookAttacks.Intersection(sideRooks).Any() {
		return false
	}
	for r, e := sideRooks.NextSet(0); e; r, e = sideRooks.NextSet(r + 1) {
		if File(int(r)) == File(int(kingSq)) { // 车将同一列
			if !p.IsAnyPieceBetweenFile(int(r), int(kingSq)) { // 中间无子隔挡
				return true
			}
		} else { // 同一行
			if !p.IsAnyPieceBetweenRank(int(r), int(kingSq)) {
				return true
			}
		}
	}
	return false
}

// IsCheck 返回是否被将.
func (p *Position) IsCheck() bool {
	// TODO
	if p.isKingCheck() {
		return true
	}

	if p.isRookCheck() {
		return true
	}

	// 检测是否被马将
	// 先判断将附近的八个马位是否有对方的马
	// 再判断是否别马腿

	// 检测是否被炮将
	// 先判断是否炮将在同一行或列

	// 检测是否被兵将
	// 先判断将附近的三个兵位是否有对方兵
	return false
}
