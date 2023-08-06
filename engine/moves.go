package engine

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/bits-and-blooms/bitset"
)

// Move 前 0-8 位表示 from，第 8-16 位表示 to, 16-20 位表示移动的棋子，
// 20-24 位表示表示吃掉的棋子.
type Move int32

const MoveEmpty = Move(0)

func toMove(from, to, movingPiece, capturedPiece int) Move {
	return Move(from ^ (to << 8) ^ (movingPiece << 16) ^ (capturedPiece << 20))
}

func (m Move) From() int {
	return int(m & 0xff)
}

func (m Move) To() int {
	return int((m >> 8) & 0xff)
}

// MovingPiece 返回移动的棋子.
func (m Move) MovingPiece() int {
	return int((m >> 16) & 0xf)
}

func (m Move) CapturedPiece() int {
	return int((m >> 20) & 0xf)
}

func (m Move) Parse() (from, to, movingPiece, capturedPiece int) {
	mi := int(m)
	return mi & 0xff, (mi >> 8) & 0xff, (mi >> 16) & 0xf, (mi >> 20) & 0xf
}

// String 返回着法字符表示.
func (m Move) String() string {
	if m == MoveEmpty {
		return "0000"
	}
	return SquareName(m.From()) + SquareName(m.To())
}

// StrToMove m.String() 的反函数.
func StrToMove(s string) Move {
	s = strings.ToLower(s)
	from, to := ParseSquare(s[0:2]), ParseSquare(s[2:4])
	return toMove(from, to, Empty, Empty)
}

func (p *Position) AllMovesCheckLegal() []int32 {
	movs := p.allMoves()
	legalMovs := []int32{}
	for _, mov := range movs {
		p.makeMove(mov)
		if !p.IsCheck(!p.IsRedMove) {
			legalMovs = append(legalMovs, int32(mov))
		}
		p.unMakeMove(mov)
	}
	return legalMovs
}

func (p *Position) AllMoves() []int32 {
	movs := p.allMoves()
	return *(*[]int32)(unsafe.Pointer(&movs))
}

func (p *Position) AllCaptureMoves() []int32 {
	movs := p.allCaptureMoves()
	return *(*[]int32)(unsafe.Pointer(&movs))
}

func checkAndAddMove(p *Position, movs *[]Move, mov Move) {
	// p.makeMove(mov)
	// if !p.IsCheck(!p.IsRedMove) {
	*movs = append(*movs, mov)
	// }
	// p.unMakeMove(mov)
}

func (p *Position) _allMoves() []Move {
	captureMovs := p.allCaptureMoves()
	notCaptureMovs := p.allNotCaptureMoves()
	return append(captureMovs, notCaptureMovs...)
}

func (p *Position) allMoves() []Move {
	var (
		ownPieces *bitset.BitSet
		oppPieces *bitset.BitSet
		allPieces = p.Red.Union(p.Black)
	)
	if p.IsRedMove {
		ownPieces, oppPieces = p.Red, p.Black
	} else {
		ownPieces, oppPieces = p.Black, p.Red
	}
	// target := ownPieces.Complement()
	// XXX 被将时可缩小 target 范围
	priorMovs := []Move{}
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
					mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove),
						MakePiece(p.WhatPiece(to), !p.IsRedMove))
					checkAndAddMove(p, &priorMovs, mov)
					break
				}
				// 不吃子
				mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
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
						mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove),
							MakePiece(p.WhatPiece(to), !p.IsRedMove))
						checkAndAddMove(p, &priorMovs, mov)
						break
					}
					break
				}
				if !afterShelf {
					// 不吃子
					mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove), Empty)
					checkAndAddMove(p, &movs, mov)
				}
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		tos := p.knightAttacksNg(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				checkAndAddMove(p, &priorMovs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 卒的着法
	pawns := p.Pawns.Intersection(ownPieces)
	for from, e := pawns.NextSet(0); e; from, e = pawns.NextSet(from + 1) {
		tos := LegalPawnMvs(int(from), p.IsRedMove)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				checkAndAddMove(p, &priorMovs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 象的着法
	bishops := p.Bishops.Intersection(ownPieces)
	for from, e := bishops.NextSet(0); e; from, e = bishops.NextSet(from + 1) {
		tos := p.LegalBishopMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				checkAndAddMove(p, &priorMovs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 士的着法
	advisors := p.Advisors.Intersection(ownPieces)
	for from, e := advisors.NextSet(0); e; from, e = advisors.NextSet(from + 1) {
		tos := LegalAdvisorMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				checkAndAddMove(p, &priorMovs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 将的着法
	kings := p.Kings.Intersection(ownPieces)
	for from, e := kings.NextSet(0); e; e = false {
		tos := LegalKingMvs[int(from)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				checkAndAddMove(p, &priorMovs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	/*
		kingBitSet := p.Kings.Intersection(ownPieces)
		kingSq, _ := kingBitSet.NextSet(0)
		tos := LegalKingMvs[int(kingSq)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(kingSq), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(kingSq), int(to), MakePiece(King, p.IsRedMove), Empty)
				movs = append(movs, mov)
			}
		}
	*/
	return append(priorMovs, movs...)
}

func (p *Position) allNotCaptureMoves() (ms []Move) {
	var (
		ownPieces *bitset.BitSet
		oppPieces *bitset.BitSet
		allPieces = p.Red.Union(p.Black)
	)
	if p.IsRedMove {
		ownPieces, oppPieces = p.Red, p.Black
	} else {
		ownPieces, oppPieces = p.Black, p.Red
	}
	// target := ownPieces.Complement()
	// XXX 被将时可缩小 target 范围
	priorMovs := []Move{}
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
				if oppPieces.Test(to) { // 对方棋子
					break
				}
				// 不吃子
				mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 炮的着法
	cannons := p.Cannons.Intersection(ownPieces)
	for from, e := cannons.NextSet(0); e; from, e = cannons.NextSet(from + 1) {
		deltas := []int{0x10, -0x10, 0x01, -0x01} // 上下左右四个方向
		for _, delta := range deltas {
			for i := uint(1); i <= 9; i++ {
				to := from + i*uint(delta)
				if !IsInBoard(to) { // 不在棋盘了
					break
				}
				if allPieces.Test(to) { // 阻挡
					break
				}
				// 不吃子
				mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		tos := p.knightAttacksNg(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if !allPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 卒的着法
	pawns := p.Pawns.Intersection(ownPieces)
	for from, e := pawns.NextSet(0); e; from, e = pawns.NextSet(from + 1) {
		tos := LegalPawnMvs(int(from), p.IsRedMove)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if !allPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 象的着法
	bishops := p.Bishops.Intersection(ownPieces)
	for from, e := bishops.NextSet(0); e; from, e = bishops.NextSet(from + 1) {
		tos := p.LegalBishopMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if !allPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 士的着法
	advisors := p.Advisors.Intersection(ownPieces)
	for from, e := advisors.NextSet(0); e; from, e = advisors.NextSet(from + 1) {
		tos := LegalAdvisorMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if !allPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	// 将的着法
	kings := p.Kings.Intersection(ownPieces)
	for from, e := kings.NextSet(0); e; e = false {
		tos := LegalKingMvs[int(from)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if !allPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove), Empty)
				checkAndAddMove(p, &movs, mov)
			}
		}
	}
	/*
		kingBitSet := p.Kings.Intersection(ownPieces)
		kingSq, _ := kingBitSet.NextSet(0)
		tos := LegalKingMvs[int(kingSq)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(kingSq), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(kingSq), int(to), MakePiece(King, p.IsRedMove), Empty)
				movs = append(movs, mov)
			}
		}
	*/
	return append(priorMovs, movs...)
}

type moveWithValue struct {
	mov Move
	val int
}

var shellSortGaps = [...]int{19, 5, 1}

func sortMovsWithValue(mvs []*moveWithValue) {
	for _, gap := range shellSortGaps {
		for i := gap; i < len(mvs); i++ {
			j, t := i, mvs[i]
			for ; j >= gap && mvs[j].val < t.val; j -= gap {
				mvs[j] = mvs[j-gap]
			}
			mvs[j] = t
		}
	}
}

func (p *Position) _allCaptureMoves() (ms []Move) {
	mvs := p.allCaptureMovesWithValue()
	sortMovsWithValue(mvs)
	for i := range mvs {
		ms = append(ms, mvs[i].mov)
	}
	return
}

func (p *Position) allCaptureMoves() (ms []Move) {
	var (
		ownPieces *bitset.BitSet
		oppPieces *bitset.BitSet
		allPieces = p.Red.Union(p.Black)
	)
	if p.IsRedMove {
		ownPieces, oppPieces = p.Red, p.Black
	} else {
		ownPieces, oppPieces = p.Black, p.Red
	}
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
					captureType := p.WhatPiece(to)
					mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove),
						MakePiece(captureType, !p.IsRedMove))
					ms = append(ms, mov)
					break
				}
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
						captureType := p.WhatPiece(to)
						mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove),
							MakePiece(captureType, !p.IsRedMove))
						ms = append(ms, mov)
						break
					}
					break
				}
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		tos := p.knightAttacksNg(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				ms = append(ms, mov)
			}
		}
	}
	// 卒的着法
	pawns := p.Pawns.Intersection(ownPieces)
	for from, e := pawns.NextSet(0); e; from, e = pawns.NextSet(from + 1) {
		tos := LegalPawnMvs(int(from), p.IsRedMove)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				ms = append(ms, mov)
			}
		}
	}
	// 象的着法
	bishops := p.Bishops.Intersection(ownPieces)
	for from, e := bishops.NextSet(0); e; from, e = bishops.NextSet(from + 1) {
		tos := p.LegalBishopMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				ms = append(ms, mov)
			}
		}
	}
	// 士的着法
	advisors := p.Advisors.Intersection(ownPieces)
	for from, e := advisors.NextSet(0); e; from, e = advisors.NextSet(from + 1) {
		tos := LegalAdvisorMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				ms = append(ms, mov)
			}
		}
	}
	// 将的着法
	kings := p.Kings.Intersection(ownPieces)
	for from, e := kings.NextSet(0); e; e = false {
		tos := LegalKingMvs[int(from)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				ms = append(ms, mov)
			}
		}
	}
	return
}

func (p *Position) allCaptureMovesWithValue() []*moveWithValue {
	var (
		ownPieces *bitset.BitSet
		oppPieces *bitset.BitSet
		allPieces = p.Red.Union(p.Black)
	)
	if p.IsRedMove {
		ownPieces, oppPieces = p.Red, p.Black
	} else {
		ownPieces, oppPieces = p.Black, p.Red
	}
	mvs := []*moveWithValue{}
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
					captureType := p.WhatPiece(to)
					mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove),
						MakePiece(captureType, !p.IsRedMove))
					mvs = append(mvs, &moveWithValue{
						mov: mov,
						val: valMap[captureType] - valMap[Rook],
					})
					break
				}
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
						captureType := p.WhatPiece(to)
						mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove),
							MakePiece(captureType, !p.IsRedMove))
						mvs = append(mvs, &moveWithValue{
							mov: mov,
							val: valMap[captureType] - valMap[Cannon],
						})
						break
					}
					break
				}
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		tos := p.knightAttacksNg(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				mvs = append(mvs, &moveWithValue{
					mov: mov,
					val: valMap[captureType] - valMap[Knight],
				})
			}
		}
	}
	// 卒的着法
	pawns := p.Pawns.Intersection(ownPieces)
	for from, e := pawns.NextSet(0); e; from, e = pawns.NextSet(from + 1) {
		tos := LegalPawnMvs(int(from), p.IsRedMove)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				mvs = append(mvs, &moveWithValue{
					mov: mov,
					val: valMap[captureType] - valMap[Pawn],
				})
			}
		}
	}
	// 象的着法
	bishops := p.Bishops.Intersection(ownPieces)
	for from, e := bishops.NextSet(0); e; from, e = bishops.NextSet(from + 1) {
		tos := p.LegalBishopMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				mvs = append(mvs, &moveWithValue{
					mov: mov,
					val: valMap[captureType] - valMap[Bishop],
				})
			}
		}
	}
	// 士的着法
	advisors := p.Advisors.Intersection(ownPieces)
	for from, e := advisors.NextSet(0); e; from, e = advisors.NextSet(from + 1) {
		tos := LegalAdvisorMvs(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				mvs = append(mvs, &moveWithValue{
					mov: mov,
					val: valMap[captureType] - valMap[Advisor],
				})
			}
		}
	}
	// 将的着法
	kings := p.Kings.Intersection(ownPieces)
	for from, e := kings.NextSet(0); e; e = false {
		tos := LegalKingMvs[int(from)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				captureType := p.WhatPiece(to)
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(captureType, !p.IsRedMove))
				mvs = append(mvs, &moveWithValue{
					mov: mov,
					val: valMap[captureType] - valMap[King],
				})
			}
		}
	}
	return mvs
}

func (p *Position) MakeMoveByDsc(dsc string) {
	if len(dsc) != 4 {
		log.Panicf("bad dsc: %s\n", dsc)
	}
	from, to := ParseSquare(dsc[0:2]), ParseSquare(dsc[2:])
	movingPiece := MakePiece(p.WhatPiece(uint(from)), p.IsRedMove)
	capturedPiece := MakePiece(p.WhatPiece(uint(to)), !p.IsRedMove)
	mov := toMove(from, to, movingPiece, capturedPiece)
	p.makeMove(mov)
}

func (p *Position) MakeMove(mov int32) {
	p.makeMove(Move(mov))
}

func (p *Position) makeMove(mov Move) {
	if mov == 0 { // 认负
		return
	}
	p.Key ^= sideKey
	fromInt, toInt, movingPiece, capturedPiece := mov.Parse()
	from, to := uint(fromInt), uint(toInt)
	movingType, isRedSide := GetPieceTypeAndSide(movingPiece)
	if p.IsRedMove != isRedSide {
		log.Printf(
			"from: 0x%x, to 0x%x, movingType: %d, capturedPiece: %d, p.IsRedMove: %v\n",
			from, to, movingType, capturedPiece, p.IsRedMove)
		log.Panicf("p.IsRedMove(%v) != isRedSide(%v)\n", p.IsRedMove, isRedSide)
	}
	if movingPiece > Pawn {
		p.Key ^= pieceSquareKey[movingPiece-2][from]
		p.Key ^= pieceSquareKey[movingPiece-2][to]
	} else {
		p.Key ^= pieceSquareKey[movingPiece-1][from]
		p.Key ^= pieceSquareKey[movingPiece-1][to]
	}

	deltaStrengthVal := 0
	if capturedPiece != Empty {
		if capturedPiece > Pawn {
			p.Key ^= pieceSquareKey[capturedPiece-2][to]
		} else {
			p.Key ^= pieceSquareKey[capturedPiece-1][to]
		}
		captureType, beCapturedSide := GetPieceTypeAndSide(capturedPiece)
		switch captureType {
		case Pawn:
			p.Pawns.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedPawnPstValue[to]
			} else {
				p.blackPstVal -= BlackPawnPstValue[to]
			}
			deltaStrengthVal = -pawnVal
		case Knight:
			p.Knights.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedKnightPstValue[to]
			} else {
				p.blackPstVal -= BlackKnightPstValue[to]
			}
			deltaStrengthVal = -knightVal
		case Rook:
			p.Rooks.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedRookPstValue[to]
			} else {
				p.blackPstVal -= BlackRookPstValue[to]
			}
			deltaStrengthVal = -rookVal
		case Cannon:
			p.Cannons.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedCannonPstValue[to]
			} else {
				p.blackPstVal -= BlackCannonPstValue[to]
			}
			deltaStrengthVal = -cannonVal
		case Bishop:
			p.Bishops.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedBishopPstValue[to]
			} else {
				p.blackPstVal -= BlackBishopPstValue[to]
			}
			deltaStrengthVal = -bishopVal
		case Advisor:
			p.Advisors.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedAdvisorPstValue[to]
			} else {
				p.blackPstVal -= BlackAdvisorPstValue[to]
			}
			deltaStrengthVal = -advisorVal
		case King:
			p.Kings.Clear(to)
			if beCapturedSide {
				p.redPstVal -= RedKingPstValue[to]
			} else {
				p.blackPstVal -= BlackKingPstValue[to]
			}
			deltaStrengthVal = -kingVal
		}
		if p.IsRedMove {
			p.Black.Clear(to)
			p.blackStrengthVal += deltaStrengthVal
		} else {
			p.Red.Clear(to)
			p.redStrengthVal += deltaStrengthVal
		}
	}
	switch movingType {
	case Pawn:
		p.Pawns.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedPawnPstValue[to] - RedPawnPstValue[from]
		} else {
			p.blackPstVal += BlackPawnPstValue[to] - BlackPawnPstValue[from]
		}
	case Knight:
		p.Knights.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedKnightPstValue[to] - RedKnightPstValue[from]
		} else {
			p.blackPstVal += BlackKnightPstValue[to] - BlackKnightPstValue[from]
		}
	case Cannon:
		p.Cannons.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedCannonPstValue[to] - RedCannonPstValue[from]
		} else {
			p.blackPstVal += BlackCannonPstValue[to] - BlackCannonPstValue[from]
		}
	case Rook:
		p.Rooks.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedRookPstValue[to] - RedRookPstValue[from]
		} else {
			p.blackPstVal += BlackRookPstValue[to] - BlackRookPstValue[from]
		}
	case Bishop:
		p.Bishops.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedBishopPstValue[to] - RedBishopPstValue[from]
		} else {
			p.blackPstVal += BlackBishopPstValue[to] - BlackBishopPstValue[from]
		}
	case Advisor:
		p.Advisors.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedAdvisorPstValue[to] - RedAdvisorPstValue[from]
		} else {
			p.blackPstVal += BlackAdvisorPstValue[to] - BlackAdvisorPstValue[from]
		}
	case King:
		p.Kings.Clear(from).Set(to)
		if isRedSide {
			p.redPstVal += RedKingPstValue[to] - RedKingPstValue[from]
		} else {
			p.blackPstVal += BlackKingPstValue[to] - BlackKingPstValue[from]
		}
	}
	if p.IsRedMove {
		p.Red.Clear(from).Set(to)
	} else {
		p.Black.Clear(from).Set(to)
	}
	p.IsRedMove = !p.IsRedMove
	/*
		// check key
		key := p.ComputeKey()
		if p.Key != key {
			log.Panicf("key: %s, p.Key: %s\n", key, p.Key)
		}
	*/
}

func (p *Position) UnMakeMove(mov int32) {
	p.unMakeMove(Move(mov))
}

func (p *Position) unMakeMove(mov Move) {
	p.Key ^= sideKey
	fromInt, toInt, movingPiece, capturedPiece := mov.Parse()
	from, to := uint(fromInt), uint(toInt)
	movingType, _ := GetPieceTypeAndSide(movingPiece)
	if movingPiece > Pawn {
		p.Key ^= pieceSquareKey[movingPiece-2][from]
		p.Key ^= pieceSquareKey[movingPiece-2][to]
	} else {
		p.Key ^= pieceSquareKey[movingPiece-1][from]
		p.Key ^= pieceSquareKey[movingPiece-1][to]
	}
	switch movingType {
	case Pawn:
		p.Pawns.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackPawnPstValue[from] - BlackPawnPstValue[to]
		} else {
			p.redPstVal += RedPawnPstValue[from] - RedPawnPstValue[to]
		}
	case Knight:
		p.Knights.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackKnightPstValue[from] - BlackKnightPstValue[to]
		} else {
			p.redPstVal += RedKnightPstValue[from] - RedKnightPstValue[to]
		}
	case Cannon:
		p.Cannons.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackCannonPstValue[from] - BlackCannonPstValue[to]
		} else {
			p.redPstVal += RedCannonPstValue[from] - RedCannonPstValue[to]
		}
	case Rook:
		p.Rooks.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackRookPstValue[from] - BlackRookPstValue[to]
		} else {
			p.redPstVal += RedRookPstValue[from] - RedRookPstValue[to]
		}
	case Bishop:
		p.Bishops.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackBishopPstValue[from] - BlackBishopPstValue[to]
		} else {
			p.redPstVal += RedBishopPstValue[from] - RedBishopPstValue[to]
		}
	case Advisor:
		p.Advisors.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackAdvisorPstValue[from] - BlackAdvisorPstValue[to]
		} else {
			p.redPstVal += RedAdvisorPstValue[from] - RedAdvisorPstValue[to]
		}
	case King:
		p.Kings.Clear(to).Set(from)
		if p.IsRedMove {
			p.blackPstVal += BlackKingPstValue[from] - BlackKingPstValue[to]
		} else {
			p.redPstVal += RedKingPstValue[from] - RedKingPstValue[to]
		}
	}
	if p.IsRedMove {
		p.Black.Clear(to).Set(from)
	} else {
		p.Red.Clear(to).Set(from)
	}
	deltaStrengthVal := 0
	if capturedPiece != Empty {
		if capturedPiece > Pawn {
			p.Key ^= pieceSquareKey[capturedPiece-2][to]
		} else {
			p.Key ^= pieceSquareKey[capturedPiece-1][to]
		}
		captureType, beCapturedSide := GetPieceTypeAndSide(capturedPiece)
		switch captureType {
		case Pawn:
			p.Pawns.Set(to)
			if beCapturedSide {
				p.redPstVal += RedPawnPstValue[to]
			} else {
				p.blackPstVal += BlackPawnPstValue[to]
			}
			deltaStrengthVal = pawnVal
		case Knight:
			p.Knights.Set(to)
			if beCapturedSide {
				p.redPstVal += RedKnightPstValue[to]
			} else {
				p.blackPstVal += BlackKnightPstValue[to]
			}
			deltaStrengthVal = knightVal
		case Rook:
			p.Rooks.Set(to)
			if beCapturedSide {
				p.redPstVal += RedRookPstValue[to]
			} else {
				p.blackPstVal += BlackRookPstValue[to]
			}
			deltaStrengthVal = rookVal
		case Cannon:
			p.Cannons.Set(to)
			if beCapturedSide {
				p.redPstVal += RedCannonPstValue[to]
			} else {
				p.blackPstVal += BlackCannonPstValue[to]
			}
			deltaStrengthVal = cannonVal
		case Bishop:
			p.Bishops.Set(to)
			if beCapturedSide {
				p.redPstVal += RedBishopPstValue[to]
			} else {
				p.blackPstVal += BlackBishopPstValue[to]
			}
			deltaStrengthVal = bishopVal
		case Advisor:
			p.Advisors.Set(to)
			if beCapturedSide {
				p.redPstVal += RedAdvisorPstValue[to]
			} else {
				p.blackPstVal += BlackAdvisorPstValue[to]
			}
			deltaStrengthVal = advisorVal
		case King:
			p.Kings.Set(to)
			if beCapturedSide {
				p.redPstVal += RedKingPstValue[to]
			} else {
				p.blackPstVal += BlackKingPstValue[to]
			}
			deltaStrengthVal = kingVal
		}
		if p.IsRedMove {
			p.Red.Set(to)
			p.redStrengthVal += deltaStrengthVal
		} else {
			p.Black.Set(to)
			p.blackStrengthVal += deltaStrengthVal
		}
	}
	p.IsRedMove = !p.IsRedMove
	/*
		// check key
		key := p.ComputeKey()
		if p.Key != key {
			log.Panicf("key: %s, p.Key: %s\n", key, p.Key)
		}
	*/
}

func (p *Position) Perft(depth uint) (nodes int) {
	startT := time.Now()
	nodes = p.perft(depth, true)
	fmt.Printf("depth: %d, nodes: %d\ntime: %v\n", depth, nodes, time.Since(startT))
	return nodes
}

func (p *Position) perft(depth uint, root bool) (nodes int) {
	moves := p.AllMovesCheckLegal()
	if depth <= 1 {
		if root {
			for i, move := range moves {
				fmt.Printf("%3d: %s: 1\n", i+1, Move(move).String())
			}
		}
		return len(moves)
	}
	for _, move := range moves {
		p.MakeMove(move)
		cnt := p.perft(depth-1, false)
		nodes += cnt
		p.UnMakeMove(move)

		if root {
			fmt.Printf("%s: %d\n", Move(move).String(), cnt)
		}
	}
	return nodes
}
