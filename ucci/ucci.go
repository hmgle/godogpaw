package ucci

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hmgle/godogpaw/engine"
	"github.com/sirupsen/logrus"
)

type Protocol struct {
	cmds map[string]func(p *Protocol, args []string)
}

func NewProtocol() *Protocol {
	p := &Protocol{}
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

func sendLine(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	logrus.WithFields(logrus.Fields{
		"direction": "out",
		"payload":   line,
	}).Debug("ucci reply")
	fmt.Println(line)
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
	bestMov := enginePosition.SearchPosition(depth)
	logrus.WithFields(logrus.Fields{
		"direction": "out",
		"command":   "bestmove",
		"move":      engine.Move2Str(bestMov),
		"depth":     depth,
	}).Debug("computed move")
	sendLine("bestmove %s", engine.Move2Str(bestMov))
}

func banmovesCmd(p *Protocol, args []string) {
	// TODO
}

const initFen = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

var enginePosition engine.PositionNG

// 格式：position {fen <FEN串> | startpos} [moves <后续着法列表>]
func positionCmd(p *Protocol, args []string) {
	if len(args) == 0 {
		log.Printf("position command missing arguments")
		return
	}
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
	enginePosition.Set(fen)
	if movesIndex >= 0 {
		for _, mv := range args[movesIndex+1:] {
			move, err := engine.ParseUCIMove(&enginePosition, mv)
			if err != nil {
				sendLine("info string invalid move %s: %v", mv, err)
				return
			}
			var st engine.StateInfo
			enginePosition.DoMove(move, &st)
		}
	}
	isOk := enginePosition.PosIsOk()
	log.Printf("fen: %s, p.PosIsOk: %+v, eval: %d, red_ksq: %d, black_ksq: %d\n",
		fen, isOk, enginePosition.Evaluate(), enginePosition.KingSQ[engine.WHITE], enginePosition.KingSQ[engine.BLACK])
}

func perftCmd(p *Protocol, args []string) {
	if len(args) == 0 {
		sendLine("info string usage: perft <depth>")
		return
	}
	depth, err := strconv.Atoi(args[0])
	if err != nil || depth < 0 {
		sendLine("info string invalid depth %s", args[0])
		return
	}
	start := time.Now()
	nodes := enginePosition.Perft(uint(depth), false)
	elapsed := time.Since(start)
	nps := 0
	if elapsed > 0 {
		nps = int(float64(nodes) / elapsed.Seconds())
	}
	sendLine("info string perft depth %d nodes %d time %dms nps %d", depth, nodes, elapsed.Milliseconds(), nps)
	sendLine("perft %d", nodes)
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
	sendLine("readyok")
}

func ucciCmd(p *Protocol, args []string) {
	sendLine("ucciok")
}

func (p *Protocol) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdLine := scanner.Text()
		if cmdLine == "quit" {
			return
		}
		logrus.WithFields(logrus.Fields{
			"direction": "in",
			"payload":   cmdLine,
		}).Debug("ucci recv")
		cmdArgs := strings.Fields(cmdLine)
		cmd, ok := p.cmds[cmdArgs[0]]
		if ok {
			cmd(p, cmdArgs[1:])
		} else {
			log.Fatalln("bad cmd:", cmdLine)
		}
	}
}
