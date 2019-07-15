package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shadow1163/logger"
	"github.com/shadow1163/nes-go/step0/nes"
)

func main() {
	log := logger.NewLogger()
	flag.Parse()

	var args []string = flag.Args()

	if len(args) != 1 {
		fmt.Println("Usage: nes FILENAME.ROM")
		flag.PrintDefaults()
		os.Exit(1)
	}
	cart, err := nes.LoadCartridge(args[0])
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	_ = cart
	// log.Printf("ROM: RPG-RPM: %d x 16KB  CHR-ROM %d x 8KB Mapper: %d", len(cart.RPG))
	// log.Info(cart)
}
