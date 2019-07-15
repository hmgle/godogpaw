package ucci

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type Engine interface {
	GetInfo() (name, version, author string)
	Prepare()
	Position(fen string)
	Search(depth uint8) (movDesc string, score int)
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
	// 反馈：bestmove <最佳着法> [ponder <后台思考的猜测着法>] [draw | resign]
	bestMov, score := p.eng.Search(4)
	fmt.Printf("info depth 4 score %d pv\n", score)
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
		cmdArgs := strings.Fields(cmdLine)
		cmdName := cmdArgs[0]
		cmd, ok := p.cmds[cmdName]
		if ok {
			cmd(p, cmdArgs[1:])
		} else {
			log.Fatalln("bad cmd:", cmdLine)
		}
	}
}
