package engine

import "testing"

func benchmarkEvaluate(b *testing.B, fen string) {
	var pos PositionNG
	pos.Set(fen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pos.Evaluate()
	}
}

func BenchmarkEvaluateStartPos(b *testing.B) {
	benchmarkEvaluate(b, "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1")
}

func BenchmarkEvaluateMidgame(b *testing.B) {
	benchmarkEvaluate(b, "2bakab2/4n4/3c1r3/p1p1p1p1p/2n6/3R5/P1P1P1P1P/2N1C4/4A4/2BAK4 w - - 0 1")
}

func BenchmarkEvaluateEndgame(b *testing.B) {
	benchmarkEvaluate(b, "4k4/3R5/9/4P4/9/9/9/5K3/9/9 w - - 0 1")
}
