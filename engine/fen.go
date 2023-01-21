package engine

import (
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/bits-and-blooms/bitset"
	"github.com/sirupsen/logrus"
)

// NewPositionByFen 创建 Position.
// fen 格式：
// rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1
func NewPositionByFen(fen string) *Position {
	p := &Position{
		Pawns:    bitset.New(256),
		Cannons:  bitset.New(256),
		Rooks:    bitset.New(256),
		Knights:  bitset.New(256),
		Bishops:  bitset.New(256),
		Advisors: bitset.New(256),
		Kings:    bitset.New(256),
		Red:      bitset.New(256),
		Black:    bitset.New(256),
	}

	tokens := strings.Split(fen, " ")
	if len(tokens) < 5 {
		log.Fatalf("bad fen: %s", fen)
	}

	positions := strings.Split(tokens[0], "/")
	if len(positions) != 10 {
		log.Fatalf("bad fen: %s, pos: %s, len: %d", fen, tokens[0], len(positions))
	}
	for i, str := range positions {
		j := 0x02
		for _, ch := range str {
			if unicode.IsDigit(ch) {
				n, _ := strconv.Atoi(string(ch))
				j += n
			} else if unicode.IsLetter(ch) {
				pieceTyp, isRed := GetPieceTypeAndSide(ParsePiece(ch))
				sq := uint(0xb0 - i*0x10 + j)
				p.addPiece(sq, pieceTyp, isRed)
				j++
			}
		}
	}
	if tokens[1] != "b" {
		p.IsRedMove = true
	}
	p.initEval()
	p.Key = p.ComputeKey()
	logrus.WithFields(logrus.Fields{
		"p.redStrengthVal":   p.redStrengthVal,
		"p.blackStrengthVal": p.blackStrengthVal,
		"p.redPstVal":        p.redPstVal,
		"p.blackPstVal":      p.blackPstVal,
	}).Debugf("==== 初始化后评分")
	return p
}
