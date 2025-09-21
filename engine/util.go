package engine

import (
	"math/bits"
)

// / least_significant_square_bb() returns the bitboard of the least significant
// / square of a non-zero bitboard. It is equivalent to square_bb(lsb(bb)).
func LeastSignificantSquareBB(b Bitboard) Bitboard {
	if b.Lo > 0 {
		return From64(b.Lo & (-b.Lo))
	}
	return Bitboard{
		Lo: 0,
		Hi: b.Hi & (-b.Hi),
	}
}

func LeastSignificantSquareBB2(b Bitboard) Bitboard {
	return SquareBB[Lsb(b)]
}

// lsb() return the least significant bit in a non-zero bitboard
func Lsb(b Bitboard) Square {
	if b.Lo > 0 {
		return bits.TrailingZeros64(b.Lo)
	}
	return bits.TrailingZeros64(b.Hi) + 64
}

var (
	squareMask  = initSquareMask()
	squareMask2 = initSquareMask2()
)

// pop_lsb() finds and clears the least significant bit in a non-zero bitboard
func PopLsb(b *Bitboard) Square {
	s := Lsb(*b)
	*b = (*b).And((*b).SubWrap64(1))
	return s
}

func PopLsb2(b *Bitboard) Square {
	s := Lsb(*b)
	*b = (*b).And(squareMask2[s])
	return s
}

func PopLsbXor(b *Bitboard) Square {
	s := Lsb(*b)
	*b = (*b).Xor(squareMask[s])
	*b = squareMask[s].Xor((*b))
	return s
}

func Count[T comparable](s []T, value T) int {
	var count int
	for _, v := range s {
		if v == value {
			count++
		}
	}
	return count
}

func initSquareMask() [SQUARE_NB]Bitboard {
	var sqm [SQUARE_NB]Bitboard
	for sq := 0; sq < SQUARE_NB; sq++ {
		sqm[sq] = From64(1).Lsh(uint(sq))
	}
	return sqm
}

func initSquareMask2() [SQUARE_NB]Bitboard {
	var sqm [SQUARE_NB]Bitboard
	for sq := 0; sq < SQUARE_NB; sq++ {
		sqm[sq] = From64(1).Lsh(uint(sq + 1)).Sub64(1).Not()
	}
	return sqm
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
