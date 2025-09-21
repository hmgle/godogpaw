package main

import (
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/hmgle/godogpaw/ucci"
	"github.com/sirupsen/logrus"
)

func main() {
	defer logPanic()
	/*
		f, err := os.Create("cpu_profile")
		if err != nil {
			log.Fatal("create:", err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	*/

	ucciProtocol := ucci.NewProtocol()
	log.Printf("finish init\n")
	ucciProtocol.Run()
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	logPath := "/tmp/godogpaw-ucci.log"
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("open log file: %v", err)
	}

	mw := io.MultiWriter(os.Stderr, f)
	logrus.SetOutput(mw)
	log.SetOutput(mw)
	logrus.SetLevel(logrus.DebugLevel)
}

func logPanic() {
	if r := recover(); r != nil {
		logrus.WithFields(logrus.Fields{
			"panic": r,
			"stack": string(debug.Stack()),
		}).Error("engine panic")
		panic(r)
	}
}
