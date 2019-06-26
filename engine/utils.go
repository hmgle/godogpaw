package engine

import (
	"strings"
	"unicode"
)

// MakePiece 返回棋子的数字描述.
func MakePiece(pieceType int, isRedSide bool) int {
	if isRedSide {
		return pieceType
	}
	return pieceType + 8
}

// ParsePiece 解析棋子.
func ParsePiece(ch rune) int {
	// 小写为黑，大写为红，红为真
	isRedSide := unicode.IsUpper(ch)
	spiece := string(unicode.ToLower(ch))
	i := strings.Index("", spiece)
	if i < 0 {
		return Empty
	}
	return MakePiece(i+King, isRedSide)
}

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

func (m Move) MovingPiece() int {
	return int((m >> 16) & 7)
}

func (m Move) CapturedPiece() int {
	return int((m >> 19) & 7)
}

func GetPieceTypeAndSide(piece int) (piectType int, isRedSide bool) {
	if piece <= Pawn { // 红
		return piece, true
	}
	return piece - 7, false
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
