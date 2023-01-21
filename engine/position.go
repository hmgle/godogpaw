package engine

import (
	"fmt"
	"log"

	"github.com/bits-and-blooms/bitset"
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

	redStrengthVal   int // 红方子力价值
	blackStrengthVal int // 黑方子力价值
	redPstVal        int // 红方位置价值
	blackPstVal      int // 黑方位置价值

	// PiecesSq [256]int // 存放每个位置的棋子

	CntRed   uint
	CntBlack uint

	IsRedMove bool
	// Key 当前局面哈希
	Key uint64
}

func (p *Position) addPiece(sq uint, pieceTyp int, isRed bool) {
	switch pieceTyp {
	case Pawn:
		p.Pawns.Set(sq)
	case Cannon:
		p.Cannons.Set(sq)
	case Rook:
		p.Rooks.Set(sq)
	case Knight:
		p.Knights.Set(sq)
	case Bishop:
		p.Bishops.Set(sq)
	case Advisor:
		p.Advisors.Set(sq)
	case King:
		p.Kings.Set(sq)
	default:
		log.Fatalf("bad pieceTyp: %d, sq: %x, isRed: %v\n", pieceTyp, sq, isRed)
	}
	if isRed {
		p.Red.Set(sq)
	} else {
		p.Black.Set(sq)
	}
}

func (p *Position) GetKey() uint64 {
	return p.Key
}

func (p *Position) IsMaximizingPlayerTurn() bool {
	return p.IsRedMove
}

// WhatPiece 返回 sq 位置的棋子类型.
// XXX 可用一个 256 数组 PiecesSq 存储每个位置的棋子类型提高速度.
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
	panic(fmt.Errorf("wrong piece on 0x%x", sq))
}

// isKingCheck 返回将帅是否照面.
func (p *Position) isKingCheck() bool {
	// 是否同一列
	redKing, found := p.Kings.NextSet(0)
	if !found {
		log.Panic("not found red king")
	}
	blackKing, found := p.Kings.NextSet(redKing + 1)
	if !found {
		log.Panicf("not found black king, redKing: 0x%x", redKing)
	}
	if File(int(redKing)) != File(int(blackKing)) {
		return false
	}

	return !p.IsAnyPieceBetweenFile(int(redKing), int(blackKing))
}

// IsOnePieceBetweenFile 判断同一列的两个棋子(sq1, sq2)之间是否有且仅有一个棋子.
// sq1, sq2 必须是同一列的，即 File(sq1) == File(sq2).
func (p *Position) IsOnePieceBetweenFile(sq1, sq2 int) bool {
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
	if int(next) >= max {
		return false
	}
	next, _ = fileMask.NextSet(next + 1)
	return int(next) == max
}

// IsOnePieceBetweenRank 判断同一行的两个棋子(sq1, sq2)之间是否有且仅有一个棋子.
// sq1, sq2 必须是同一行的，即 Rank(sq1) == Rank(sq2).
func (p *Position) IsOnePieceBetweenRank(sq1, sq2 int) bool {
	rank := Rank(sq1)
	if Rank(sq2) != rank {
		log.Fatalf("sq1(0x%x) and sq2(0x%x) is not in same rank", sq1, sq2)
	}

	min, max := sq1, sq2
	if sq1 > sq2 {
		min, max = sq2, sq1
	}
	gb := p.Black.Union(p.Red)
	rankMask := gb.Intersection(RankMasks[rank])
	next, _ := rankMask.NextSet(uint(min + 1))
	if int(next) >= max {
		return false
	}
	next, _ = rankMask.NextSet(next + 1)
	return int(next) == max
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

// knightAttacks 返回马位于 sq 位置时的攻击点.
func (p *Position) knightAttacks(sq uint) *bitset.BitSet {
	mask := bitset.New(256)
	gb := p.Black.Union(p.Red)
	if gb.Test(sq + 1) {
		mask.Set(sq + 0x10 + 2)
		mask.Set(sq - 0x10 + 2)
	}
	if gb.Test(sq - 1) {
		mask.Set(sq + 0x10 - 2)
		mask.Set(sq - 0x10 - 2)
	}
	if gb.Test(sq + 0x10) {
		mask.Set(sq + 0x20 + 1)
		mask.Set(sq + 0x20 - 1)
	}
	if gb.Test(sq - 0x10) {
		mask.Set(sq - 0x20 + 1)
		mask.Set(sq - 0x20 - 1)
	}
	return KnightAttacks[int(sq)].Difference(mask)
}

func (p *Position) knightAttacksNg(sq uint) *bitset.BitSet {
	atts := bitset.New(256)
	gb := p.Black.Union(p.Red)
	if !gb.Test(sq + 1) {
		atts.Set(sq + 0x10 + 2)
		atts.Set(sq - 0x10 + 2)
	}
	if !gb.Test(sq - 1) {
		atts.Set(sq + 0x10 - 2)
		atts.Set(sq - 0x10 - 2)
	}
	if !gb.Test(sq + 0x10) {
		atts.Set(sq + 0x20 + 1)
		atts.Set(sq + 0x20 - 1)
	}
	if !gb.Test(sq - 0x10) {
		atts.Set(sq - 0x20 + 1)
		atts.Set(sq - 0x20 - 1)
	}
	atts.InPlaceIntersection(BoardMask)
	return atts
}

// isKnightCheck 返回己方马是否在将.
func (p *Position) isKnightCheck(isRedCheck bool) bool {
	var (
		kingSq uint
		selfPs *bitset.BitSet
		sidePs *bitset.BitSet
	)
	if isRedCheck {
		selfPs, sidePs = p.Black, p.Red
	} else {
		selfPs, sidePs = p.Red, p.Black
	}
	kingSq, _ = p.Kings.Intersection(sidePs).NextSet(0)
	knightAttacks := KnightAttacks[int(kingSq)]
	selfKnights := p.Knights.Intersection(selfPs)
	// 先判断将附近的八个马位是否有己方的马
	if !knightAttacks.Intersection(selfKnights).Any() {
		return false
	}
	for k, e := selfKnights.NextSet(0); e; k, e = selfKnights.NextSet(k + 1) {
		if p.knightAttacksNg(k).Test(kingSq) {
			return true
		}
	}
	return false
}

// isPawnCheck 返回己方兵是否在将.
func (p *Position) isPawnCheck(isRedCheck bool) bool {
	var (
		kingSq uint
		selfPs *bitset.BitSet
		sidePs *bitset.BitSet
	)
	if isRedCheck {
		selfPs, sidePs = p.Black, p.Red
	} else {
		selfPs, sidePs = p.Red, p.Black
	}
	kingSq, _ = p.Kings.Intersection(sidePs).NextSet(0)
	pawnAttacks := AttackKingPawnSqs[int(kingSq)]
	selfPawns := p.Pawns.Intersection(selfPs)
	return pawnAttacks.Intersection(selfPawns).Any()
}

// isCannonCheck 返回己方炮是否在将.
func (p *Position) isCannonCheck(isRedCheck bool) bool {
	var (
		kingSq uint
		selfPs *bitset.BitSet
		sidePs *bitset.BitSet
	)
	if isRedCheck {
		selfPs, sidePs = p.Black, p.Red
	} else {
		selfPs, sidePs = p.Red, p.Black
	}
	kingSq, _ = p.Kings.Intersection(sidePs).NextSet(0)
	rookAttacks := RookAttacks[int(kingSq)]
	selfCannons := p.Cannons.Intersection(selfPs)
	// 先判断是否对方帅同一行及同一列有没有己方炮
	if !rookAttacks.Intersection(selfCannons).Any() {
		return false
	}

	for c, e := selfCannons.NextSet(0); e; c, e = selfCannons.NextSet(c + 1) {
		if File(int(c)) == File(int(kingSq)) { // 炮将同一列
			if p.IsOnePieceBetweenFile(int(c), int(kingSq)) { // 中间一子隔挡
				return true
			}
		} else if Rank(int(c)) == Rank(int(kingSq)) { // 同一行
			if p.IsOnePieceBetweenRank(int(c), int(kingSq)) {
				return true
			}
		}
	}

	return false
}

// isRookCheck 返回车是否在将.
// isRedCheck = true: 返回红帅是否被黑车将.
// isRedCheck = false: 返回黑将是否被红车将.
func (p *Position) isRookCheck(isRedCheck bool) bool {
	var (
		kingSq uint
		selfPs *bitset.BitSet
		sidePs *bitset.BitSet
	)
	if isRedCheck {
		selfPs, sidePs = p.Black, p.Red
	} else {
		selfPs, sidePs = p.Red, p.Black
	}
	kingSq, _ = p.Kings.Intersection(sidePs).NextSet(0)
	rookAttacks := RookAttacks[int(kingSq)]
	selfRooks := p.Rooks.Intersection(selfPs)
	// 先判断是否对方帅同一行及同一列有没有己方车
	if !rookAttacks.Intersection(selfRooks).Any() {
		return false
	}
	for r, e := selfRooks.NextSet(0); e; r, e = selfRooks.NextSet(r + 1) {
		if File(int(r)) == File(int(kingSq)) { // 车将同一列
			if !p.IsAnyPieceBetweenFile(int(r), int(kingSq)) { // 中间无子隔挡
				return true
			}
		} else if Rank(int(r)) == Rank(int(kingSq)) { // 同一行
			if !p.IsAnyPieceBetweenRank(int(r), int(kingSq)) {
				return true
			}
		}
	}
	return false
}

// IsCheck 返回是否将.
// isRedCheck = true: 返回红帅是否被将.
// isRedCheck = false: 返回黑将是否被将.
func (p *Position) IsCheck(isRedCheck bool) bool {
	if p.isKingCheck() {
		return true
	}
	if p.isRookCheck(isRedCheck) {
		return true
	}

	if p.isCannonCheck(isRedCheck) {
		return true
	}
	if p.isKnightCheck(isRedCheck) {
		return true
	}

	if p.isPawnCheck(isRedCheck) {
		return true
	}
	return false
}

// AllPieces 返回所有棋子.
func (p *Position) AllPieces() *bitset.BitSet {
	return p.Red.Union(p.Black)
}

// LegalBishopMvs 返回 sq 这个位置象的合法着法位置.
func (p *Position) LegalBishopMvs(sq uint) *bitset.BitSet {
	allPieces := p.AllPieces()
	mvsBs := bitset.New(256)
	if sq < 0x2f { // 底相
		if !allPieces.Test(sq + 0x10 + 0x01) {
			mvsBs.Set(sq + 0x20 + 0x02)
		}
		if !allPieces.Test(sq + 0x10 - 0x01) {
			mvsBs.Set(sq + 0x20 - 0x02)
		}
	} else if sq < 0x50 { // 中相
		if !allPieces.Test(sq + 0x10 + 0x01) {
			mvsBs.Set(sq + 0x20 + 0x02)
		}
		if !allPieces.Test(sq + 0x10 - 0x01) {
			mvsBs.Set(sq + 0x20 - 0x02)
		}
		if !allPieces.Test(sq - 0x10 + 0x01) {
			mvsBs.Set(sq - 0x20 + 0x02)
		}
		if !allPieces.Test(sq - 0x10 - 0x01) {
			mvsBs.Set(sq - 0x20 - 0x02)
		}
	} else if sq < 0x6f { // 高相不能过河
		if !allPieces.Test(sq - 0x10 + 0x01) {
			mvsBs.Set(sq - 0x20 + 0x02)
		}
		if !allPieces.Test(sq - 0x10 - 0x01) {
			mvsBs.Set(sq - 0x20 - 0x02)
		}
	} else if sq < 0x80 { // 高象不能过河
		if !allPieces.Test(sq + 0x10 + 0x01) {
			mvsBs.Set(sq + 0x20 + 0x02)
		}
		if !allPieces.Test(sq + 0x10 - 0x01) {
			mvsBs.Set(sq + 0x20 - 0x02)
		}
	} else if sq < 0xaf { // 中象
		if !allPieces.Test(sq + 0x10 + 0x01) {
			mvsBs.Set(sq + 0x20 + 0x02)
		}
		if !allPieces.Test(sq + 0x10 - 0x01) {
			mvsBs.Set(sq + 0x20 - 0x02)
		}
		if !allPieces.Test(sq - 0x10 + 0x01) {
			mvsBs.Set(sq - 0x20 + 0x02)
		}
		if !allPieces.Test(sq - 0x10 - 0x01) {
			mvsBs.Set(sq - 0x20 - 0x02)
		}
	} else { // 底象
		if !allPieces.Test(sq - 0x10 + 0x01) {
			mvsBs.Set(sq - 0x20 + 0x02)
		}
		if !allPieces.Test(sq - 0x10 - 0x01) {
			mvsBs.Set(sq - 0x20 - 0x02)
		}
	}
	mvsBs.InPlaceIntersection(BoardMask)
	return mvsBs
}
