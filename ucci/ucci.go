package ucci

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Engine interface {
	GetInfo() (name, version, author string)
	Prepare()
	Position(fen string)
	Move(movDsc string)
	Search(depth uint8) (movDesc string, score int)
	Perft(depth uint) (nodes int)
}

type Protocol struct {
	cmds map[string]func(p *Protocol, args []string)
	eng  Engine
}

func NewProtocol(e Engine) *Protocol {
	p := &Protocol{
		eng: e,
	}
	p.cmds = map[string]func(p *Protocol, args []string){
		"ucci":      ucciCmd,
		"isready":   isReadyCmd,
		"setoption": setOptionCmd,
		"position":  positionCmd,
		"banmoves":  banmovesCmd,
		"go":        goCmd,
		"ponderhit": ponderhitCmd,
		"stop":      stopCmd,
		"perft":     perftCmd,
	}
	return p
}

func stopCmd(p *Protocol, args []string) {
	// TODO
}

func ponderhitCmd(p *Protocol, args []string) {
	// TODO
}

func goCmd(p *Protocol, args []string) {
	// TODO
	// go [ponder | draw] <思考模式>
	// 反馈：bestmove <最佳着法> [ponder <后台思考的猜测着法>] [draw | resign]
	depth := uint8(4)
	if len(args) > 1 && args[0] == "depth" {
		newDepth, err := strconv.Atoi(args[1])
		if err != nil {
			log.Panic(err)
		}
		depth = uint8(newDepth)
	}
	// XXX DEBUG
	// depth = uint8(6)
	bestMov, score := p.eng.Search(depth)
	fmt.Printf("info depth %d score %d pv\n", depth, score)
	logrus.WithFields(logrus.Fields{
		"bestmove": bestMov,
		"depth":    depth,
		"score":    score,
	}).Debugf("返回最佳着法")
	fmt.Printf("bestmove %s\n", bestMov)
}

func banmovesCmd(p *Protocol, args []string) {
	// TODO
}

const initFen = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

// 格式：position {fen <FEN串> | startpos} [moves <后续着法列表>]
func positionCmd(p *Protocol, args []string) {
	var fen string
	movesIndex := findIndexString(args, "moves")
	if args[0] == "startpos" {
		fen = initFen
	} else if args[0] == "fen" {
		if movesIndex == -1 {
			fen = strings.Join(args[1:], " ")
		} else {
			fen = strings.Join(args[1:movesIndex], " ")
		}
	} else {
		log.Fatalf("bad fen: %v", args)
	}
	p.eng.Position(fen)
	if movesIndex >= 0 {
		for _, dscMov := range args[movesIndex+1:] {
			p.eng.Move(dscMov)
		}
	}
}

func findIndexString(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func setOptionCmd(p *Protocol, args []string) {
	// TODO
}

func isReadyCmd(p *Protocol, args []string) {
	p.eng.Prepare()
	fmt.Println("readyok")
}

func perftCmd(p *Protocol, args []string) {
	depth := uint(1)
	if len(args) > 0 {
		newDepth, err := strconv.Atoi(args[0])
		if err != nil {
			log.Panic(err)
		}
		depth = uint(newDepth)
	}
	p.eng.Perft(depth)
}

func ucciCmd(p *Protocol, args []string) {
	name, version, author := p.eng.GetInfo()
	fmt.Printf("id name %s %s\n", name, version)
	fmt.Printf("id author %s\n", author)
	fmt.Println("ucciok")
}

func (p *Protocol) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdLine := scanner.Text()
		if cmdLine == "quit" {
			return
		}
		logrus.WithFields(logrus.Fields{
			"cmd": cmdLine,
		}).Debug("")
		cmdArgs := strings.Fields(cmdLine)
		cmd, ok := p.cmds[cmdArgs[0]]
		if ok {
			cmd(p, cmdArgs[1:])
		} else {
			log.Fatalln("bad cmd:", cmdLine)
		}
	}
}
