package engine

import "testing"

func BenchmarkKnightsMovs(b *testing.B) {
	p := NewPositionByFen("3akab2/4n3n/r3b4/CR1N2N1p/2pP5/9/P8/4C3B/4A2c1/2c1KA3 b - - 0 1")
	b.Run("knightAttacks", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = p.knightAttacks(0x67)
		}
	})
	b.Run("knightAttacksNg", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = p.knightAttacksNg(0x67)
		}
	})
}
