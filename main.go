package main

import (
	"log"
	"os"
	"runtime/pprof"

	hooks "github.com/git-hulk/logrus-hooks"
	"github.com/hmgle/godogpaw/engine"
	"github.com/hmgle/godogpaw/ucci"
	"github.com/sirupsen/logrus"
)

func main() {
	cpuProfile, _ := os.Create("cpu_profile")
	pprof.StartCPUProfile(cpuProfile)
	defer pprof.StopCPUProfile()

	ucciProtocol := ucci.NewProtocol(&engine.Engine{})
	ucciProtocol.Run()
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	rotateHook, err := hooks.NewRotateHook(logrus.StandardLogger(), ".", "godogpaw")
	if err != nil {
		log.Fatal(err)
	}
	logrus.AddHook(rotateHook)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(hooks.NewSourceHook(logrus.DebugLevel))
}
