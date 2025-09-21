package engine

import (
	"math/rand"
	"testing"
)

func TestLeastSignificantSquareBB(t *testing.T) {
	for i := 0; i < 100; i++ {
		b := Bitboard{
			Lo: uint64(rand.Int63()),
			Hi: uint64(rand.Int63()),
		}
		b1 := LeastSignificantSquareBB(b)
		b2 := LeastSignificantSquareBB2(b)
		if b1 != b2 {
			t.Errorf("b1: %v, b2: %v\n", b1, b2)
		}
	}
}

func TestRookAttackMap(t *testing.T) {
	for s := SQ_A0; s <= SQ_I9; s++ {
		attackMap := RookAttackMap[s]
		for k, v := range attackMap {
			a := AttacksBB(ROOK, s, k)
			if a != v {
				t.Errorf("s: %v, a: %v, v: %v, mask: %v\n", s, a, v, k)
				return
			}
		}
	}
}

func BenchmarkRookAttackMap(b *testing.B) {
	b.Run("MagicAttack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			/*
				for s := SQ_A0; s <= SQ_I9; s++ {
					attackMap := RookAttackMap[s]
					for k := range attackMap {
						_ = AttacksBB(ROOK, s, k)
					}
				}
			*/
			_ = AttacksBB(ROOK, SQ_B3, Bitboard{})
		}
	})
	b.Run("MapAttack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			/*
				for s := SQ_A0; s <= SQ_I9; s++ {
					attackMap := RookAttackMap[s]
					for k := range attackMap {
						_ = attackMap[k]
					}
				}
			*/
			_ = RookAttackMap[SQ_B3][Bitboard{}]
		}
	})

	// go test -bench=RookAttackMap -benchmem -run=none
	// goos: linux
	// goarch: amd64
	// pkg: github.com/hmgle/godogpaw/engine
	// cpu: Intel(R) Core(TM) i5-5200U CPU @ 2.20GHz
	// BenchmarkRookAttackMap/MagicAttack-4            168354465                7.206 ns/op           0 B/op          0 allocs/op
	// BenchmarkRookAttackMap/MapAttack-4              67815272                16.41 ns/op            0 B/op          0 allocs/op
	// PASS
	// ok      github.com/hmgle/godogpaw/engine        7.220s
}

func BenchmarkLeast(b *testing.B) {
	d := Bitboard{
		Lo: uint64(rand.Int63()),
		Hi: uint64(rand.Int63()),
	}
	b.Run("LeastSignificantSquareBB", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			LeastSignificantSquareBB(d)
		}
	})
	b.Run("LeastSignificantSquareBB2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			LeastSignificantSquareBB2(d)
		}
	})
}

func BenchmarkPopLsb(b *testing.B) {
	x := FromHL((1<<63)-1, 1<<63-1)
	b.Run("PooLsp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if x.IsNotZero() {
				PopLsb(&x)
			}
		}
	})
	b.Run("PooLspXor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if x.IsNotZero() {
				PopLsbXor(&x)
			}
		}
	})
	b.Run("PooLspAnd", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if x.IsNotZero() {
				PopLsb2(&x)
			}
		}
	})
}
