package main

import (
	"log"

	hooks "github.com/git-hulk/logrus-hooks"
	"github.com/sirupsen/logrus"
)

func main() {
	// WIP
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
