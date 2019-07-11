package engine

import (
	"strings"

	"github.com/willf/bitset"
)

// Move 前 0-8 位表示 from，第 8-16 位表示 to, 16-19 位表示移动的棋子，
// 19-21 位表示表示吃掉的棋子.
type Move int32

const MoveEmpty = Move(0)

func MakeMove(from, to, movingPiece, capturedPiece int) Move {
	return Move(from ^ (to << 8) ^ (movingPiece << 16) ^ (capturedPiece << 19))
}

func (m Move) From() int {
	return int(m & 0xff)
}

func (m Move) To() int {
	return int((m >> 8) & 0xff)
}

// MovingPiece 返回移动的棋子.
func (m Move) MovingPiece() int {
	return int((m >> 16) & 7)
}

func (m Move) CapturedPiece() int {
	return int((m >> 19) & 7)
}

// String 返回着法字符表示.
func (m Move) String() string {
	if m == MoveEmpty {
		return "0000"
	}
	return SquareName(m.From()) + SquareName(m.To())
}

// ParseMove m.String() 的反函数.
func ParseMove(s string) Move {
	s = strings.ToLower(s)
	from, to := ParseSquare(s[0:2]), ParseSquare(s[2:4])
	return MakeMove(from, to, Empty, Empty)
}

func GenAllMoves(p *Position) []Move {
	var (
		ownPieces *bitset.BitSet
		oppPieces *bitset.BitSet
		allPieces *bitset.BitSet = p.Red.Union(p.Black)
	)
	if p.IsRedMove {
		ownPieces, oppPieces = p.Red, p.Black
	} else {
		ownPieces, oppPieces = p.Black, p.Red
	}
	// target := ownPieces.Complement()
	// XXX 被将时可缩小 target 范围
	// TODO
	movs := []Move{}
	// 车的着法
	rooks := p.Rooks.Intersection(ownPieces)
	for from, e := rooks.NextSet(0); e; from, e = rooks.NextSet(from + 1) {
		deltas := []int{0x10, -0x10, 0x01, -0x01} // 上下左右四个方向
		for _, delta := range deltas {
			for i := uint(1); i <= 9; i++ {
				to := from + i*uint(delta)
				if ownPieces.Test(to) || !IsInBoard(to) { // 遇到自己棋子或不在棋盘了
					break
				}
				if oppPieces.Test(to) { // 吃子
					mov := MakeMove(int(from), int(to), MakePiece(Rook, p.IsRedMove),
						MakePiece(p.WhatPiece(to), !p.IsRedMove))
					movs = append(movs, mov)
					break
				}
				// 不吃子
				mov := MakeMove(int(from), int(to), MakePiece(Rook, p.IsRedMove), Empty)
				movs = append(movs, mov)
			}
		}
	}
	// 炮的着法
	cannons := p.Cannons.Intersection(ownPieces)
	for from, e := cannons.NextSet(0); e; from, e = cannons.NextSet(from + 1) {
		deltas := []int{0x10, -0x10, 0x01, -0x01} // 上下左右四个方向
		for _, delta := range deltas {
			afterShelf := false // 炮是否翻过架子
			for i := uint(1); i <= 9; i++ {
				to := from + i*uint(delta)
				if !IsInBoard(to) { // 不在棋盘了
					break
				}
				if allPieces.Test(to) { // 阻挡
					if !afterShelf {
						afterShelf = true
						continue
					}
					// 翻过了炮架，判断能否吃子
					if oppPieces.Test(to) { // 对方棋子，可吃
						mov := MakeMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove),
							MakePiece(p.WhatPiece(to), !p.IsRedMove))
						movs = append(movs, mov)
						break
					}
					break
				}
				if !afterShelf {
					// 不吃子
					mov := MakeMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove), Empty)
					movs = append(movs, mov)
				}
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		// TODO
	}
	return movs
}
