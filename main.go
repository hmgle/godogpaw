package main

import (
	"log"

	"github.com/hmgle/godogpaw/engine"
	"github.com/hmgle/godogpaw/ucci"
)

func main() {
	ucciProtocol := ucci.NewProtocol(&engine.Engine{})
	ucciProtocol.Run()
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}
