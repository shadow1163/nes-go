package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shadow1163/logger"
	"github.com/shadow1163/nes-go/step2/nes"
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
	cpu := nes.NewCPU(cart)
	log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFA)))
	log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFC)))
	log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFE)))
	log.Info(fmt.Sprintf("%2x", cpu.Read(0x6000)))
	cpu.PC = cpu.Read16(0xFFFA)
	cpu.PrintInstruction()
	cpu.PC = cpu.Read16(0xFFFC)
	cpu.PrintInstruction()
	cpu.PC = cpu.Read16(0xFFFE)
	cpu.PrintInstruction()
}
