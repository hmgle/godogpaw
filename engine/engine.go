package engine

import (
	"github.com/hmgle/godogpaw/alphabetasearch"
)

type Engine struct {
	p *Position
}

func (e *Engine) GetInfo() (name, version, author string) {
	return "godogpaw", "0.1", "hmgle"
}

func (e *Engine) Prepare() {
}

func (e *Engine) Position(fen string) {
	e.p = NewPositionByFen(fen)
}

func (e *Engine) Search(depth uint8) (movDesc string, score int) {
	bestMov, score := alphabetasearch.AlphaBetaSearch(e.p, depth, -2000, 2000)
	// mov := Move(bestMov)
	// from, to := mov.From(), mov.To()
	return Move(bestMov).String(), score
}
