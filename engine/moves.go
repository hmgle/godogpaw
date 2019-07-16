package engine

import (
	"log"
	"strings"
	"unsafe"

	"github.com/willf/bitset"
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

func (p *Position) AllMoves() []int32 {
	movs := p.allMoves()
	return *(*[]int32)(unsafe.Pointer(&movs))
}

func (p *Position) allMoves() []Move {
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
					movs = append(movs, mov)
					break
				}
				// 不吃子
				mov := toMove(int(from), int(to), MakePiece(Rook, p.IsRedMove), Empty)
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
						mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove),
							MakePiece(p.WhatPiece(to), !p.IsRedMove))
						movs = append(movs, mov)
						break
					}
					break
				}
				if !afterShelf {
					// 不吃子
					mov := toMove(int(from), int(to), MakePiece(Cannon, p.IsRedMove), Empty)
					movs = append(movs, mov)
				}
			}
		}
	}
	// 马的着法
	knights := p.Knights.Intersection(ownPieces)
	for from, e := knights.NextSet(0); e; from, e = knights.NextSet(from + 1) {
		tos := p.knightAttacks(from)
		for to, e2 := tos.NextSet(0); e2; to, e2 = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Knight, p.IsRedMove), Empty)
				movs = append(movs, mov)
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
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Pawn, p.IsRedMove), Empty)
				movs = append(movs, mov)
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
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Bishop, p.IsRedMove), Empty)
				movs = append(movs, mov)
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
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(Advisor, p.IsRedMove), Empty)
				movs = append(movs, mov)
			}
		}
	}
	// 将的着法
	kings := p.Advisors.Intersection(ownPieces)
	for from, e := kings.NextSet(0); e; e = false {
		tos := LegalKingMvs[int(from)]
		for to, e := tos.NextSet(0); e; to, e = tos.NextSet(to + 1) {
			if oppPieces.Test(to) { // 吃子
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove),
					MakePiece(p.WhatPiece(to), !p.IsRedMove))
				movs = append(movs, mov)
			} else if !ownPieces.Test(to) { // 不吃子
				mov := toMove(int(from), int(to), MakePiece(King, p.IsRedMove), Empty)
				movs = append(movs, mov)
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
	return movs
}

func (p *Position) checkLegalPos(movInt32 int32) {
	all1 := p.Red.Union(p.Black)
	all2 := p.Pawns.Union(p.Cannons).Union(p.Rooks).Union(p.Knights).
		Union(p.Bishops).Union(p.Advisors).Union(p.Kings)
	if all1.DifferenceCardinality(all2) == 0 {
		// ok
		return
	}
	mov := Move(movInt32)
	fromInt, toInt, movingPiece, capturedPiece := mov.Parse()
	from, to := uint(fromInt), uint(toInt)
	// movingType, isRedSide := GetPieceTypeAndSide(movingPiece)
	log.Printf("mov: 0b%025b, %d, from: 0x%x, to: 0x%x\n", mov, mov, from, to)
	log.Printf("movingPiece: %d, capturedPiece: %d\n", movingPiece, capturedPiece)
	log.Println("start===========================================")
	log.Printf("    red:\t%s\n", p.Red.StringHex())
	log.Printf("  black:\t%s\n", p.Black.StringHex())
	log.Printf("   pawn:\t%s\n", p.Pawns.StringHex())
	log.Printf("   rook:\t%s\n", p.Rooks.StringHex())
	log.Printf(" cannon:\t%s\n", p.Cannons.StringHex())
	log.Printf(" knight:\t%s\n", p.Knights.StringHex())
	log.Printf(" bishop:\t%s\n", p.Bishops.StringHex())
	log.Printf("advisor:\t%s\n", p.Advisors.StringHex())
	log.Printf("   king:\t%s\n", p.Kings.StringHex())
	log.Printf("   all1:\t%s\n", all1.StringHex())
	log.Printf("   all2:\t%s\n", all2.StringHex())
	log.Println("end===========================================")
	log.Panic("")
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
	// p.checkLegalPos(mov)
	p.makeMove(Move(mov))
	// p.checkLegalPos(mov)
}

func (p *Position) makeMove(mov Move) {
	fromInt, toInt, movingPiece, capturedPiece := mov.Parse()
	from, to := uint(fromInt), uint(toInt)
	movingType, isRedSide := GetPieceTypeAndSide(movingPiece)
	if p.IsRedMove != isRedSide {
		log.Printf("from: 0x%x, to 0x%x, movingType: %d, capturedPiece: %d, p.IsRedMove: %v\n", from, to, movingType, capturedPiece, p.IsRedMove)
		log.Panicf("p.IsRedMove(%v) != isRedSide(%v)\n", p.IsRedMove, isRedSide)
	}
	if capturedPiece != Empty {
		captureType, isRed := GetPieceTypeAndSide(capturedPiece)
		if p.IsRedMove == isRed {
			log.Panicf("p.IsRedMove(%v) == isRed(%v)\n", p.IsRedMove, isRed)
		}
		switch captureType {
		case Pawn:
			p.Pawns.Clear(to)
		case Knight:
			p.Knights.Clear(to)
		case Rook:
			p.Rooks.Clear(to)
		case Cannon:
			p.Cannons.Clear(to)
		case Bishop:
			p.Bishops.Clear(to)
		case Advisor:
			p.Advisors.Clear(to)
		case King:
			p.Kings.Clear(to)
		}
		if p.IsRedMove {
			p.Black.Clear(to)
		} else {
			p.Red.Clear(to)
		}
	}
	switch movingType {
	case Pawn:
		p.Pawns.Clear(from).Set(to)
	case Knight:
		p.Knights.Clear(from).Set(to)
	case Cannon:
		p.Cannons.Clear(from).Set(to)
	case Rook:
		p.Rooks.Clear(from).Set(to)
	case Bishop:
		p.Bishops.Clear(from).Set(to)
	case Advisor:
		p.Advisors.Clear(from).Set(to)
	case King:
		p.Kings.Clear(from).Set(to)
	}
	if p.IsRedMove {
		p.Red.Clear(from).Set(to)
	} else {
		p.Black.Clear(from).Set(to)
	}
	p.IsRedMove = !p.IsRedMove
}

func (p *Position) UnMakeMove(mov int32) {
	// p.checkLegalPos(mov)
	p.unMakeMove(Move(mov))
	// p.checkLegalPos(mov)
}

func (p *Position) unMakeMove(mov Move) {
	fromInt, toInt, movingPiece, capturedPiece := mov.Parse()
	from, to := uint(fromInt), uint(toInt)
	movingType, _ := GetPieceTypeAndSide(movingPiece)
	switch movingType {
	case Pawn:
		p.Pawns.Clear(to).Set(from)
	case Knight:
		p.Knights.Clear(to).Set(from)
	case Cannon:
		p.Cannons.Clear(to).Set(from)
	case Rook:
		p.Rooks.Clear(to).Set(from)
	case Bishop:
		p.Bishops.Clear(to).Set(from)
	case Advisor:
		p.Advisors.Clear(to).Set(from)
	case King:
		p.Kings.Clear(to).Set(from)
	}
	if p.IsRedMove {
		p.Black.Clear(to).Set(from)
	} else {
		p.Red.Clear(to).Set(from)
	}
	if capturedPiece != Empty {
		captureType, _ := GetPieceTypeAndSide(capturedPiece)
		switch captureType {
		case Pawn:
			p.Pawns.Set(to)
		case Knight:
			p.Knights.Set(to)
		case Rook:
			p.Rooks.Set(to)
		case Cannon:
			p.Cannons.Set(to)
		case Bishop:
			p.Bishops.Set(to)
		case Advisor:
			p.Advisors.Set(to)
		case King:
			p.Kings.Set(to)
		}
		if p.IsRedMove {
			p.Red.Set(to)
		} else {
			p.Black.Set(to)
		}
	}
	p.IsRedMove = !p.IsRedMove
}
