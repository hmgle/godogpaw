package engine

import (
	"github.com/hmgle/godogpaw/alphabetasearch"
	"github.com/sirupsen/logrus"
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

func (e *Engine) Move(movDsc string) {
	e.p.MakeMoveByDsc(movDsc)
}

func (e *Engine) Search(depth uint8) (movDesc string, score int) {
	val := e.p.Evaluate()
	logrus.WithFields(logrus.Fields{
		"val": val,
	}).Debugf("搜索前局面估值")
	bestMov, score := alphabetasearch.AlphaBetaSearch(e.p, depth, -kingVal, kingVal)
	// mov := Move(bestMov)
	// from, to := mov.From(), mov.To()
	{ // XXX DEBUG
		e.p.MakeMove(bestMov)
		val = e.p.Evaluate()
		logrus.WithFields(logrus.Fields{
			"val":              val,
			"redStrengthVal":   e.p.redStrengthVal,
			"blackStrengthVal": e.p.blackStrengthVal,
			"redPstVal":        e.p.redPstVal,
			"blackPstVal":      e.p.blackPstVal,
		}).Debugf("搜索后执行着法后的局面估值")
		e.p.UnMakeMove(bestMov)
	}
	return Move(bestMov).String(), score
}
