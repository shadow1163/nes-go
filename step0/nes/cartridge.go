package nes

import (
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/shadow1163/logger"
)

var log = logger.NewLogger()

// MirrorType mirror type
type MirrorType int

const (
	horizontal MirrorType = iota
	vertical
	singleLow
	singleHigh
	fourScreen
)

// Cartridge nes cartidge
type Cartridge struct {
	Mirror MirrorType
	// Mapper Mapper
	PRG  [][]byte // [bank][byte], 16k banks.
	CHR  [][]byte // [bank][byte], 8k banks.
	SRAM [][]byte // [bank][byte], 8k banks.
}

// NewCartridge new a cartridge
func NewCartridge(numPRGBanks int, numCHRBanks int, numSRAMBanks int) *Cartridge {
	cart := &Cartridge{}
	if numCHRBanks == 0 {
		numCHRBanks = 1
	}
	if numSRAMBanks == 0 {
		numSRAMBanks = 1
	}

	cart.PRG = make([][]byte, numPRGBanks)
	cart.CHR = make([][]byte, numCHRBanks)
	cart.SRAM = make([][]byte, numSRAMBanks)

	for i := range cart.PRG {
		cart.PRG[i] = make([]byte, 16384)
	}

	for i := range cart.CHR {
		cart.CHR[i] = make([]byte, 8192)
	}

	for i := range cart.SRAM {
		cart.SRAM[i] = make([]byte, 8192)
	}

	return cart
}

// LoadCartridge open and read an iNES format ROM file
func LoadCartridge(filename string) (*Cartridge, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	header := iNESHeader{}
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}
	if header.Magic != [3]byte{'N', 'E', 'S'} {
		return nil, errors.New("not a valid nes file")
	}
	if header.Format != 0x1a {
		return nil, errors.New("unsupported iNES format type")
	}
	if header.NumSRAMBanks == 0 {
		log.Debug("No SRAM Bank, set it.")
		header.NumSRAMBanks = 1
	}

	cart := NewCartridge(int(header.NumPRGBanks), int(header.NumCHRBanks), int(header.NumSRAMBanks))
	if header.Control1&0x02 != 0 {
		cart.Mirror = fourScreen
	} else if header.Control1&0x01 != 0 {
		cart.Mirror = vertical
	} else {
		cart.Mirror = horizontal
	}
	hasTrainer := header.Control1&0x04 != 0
	if hasTrainer {
		buf := make([]byte, 512)
		if _, err := io.ReadFull(file, buf); err != nil {
			return nil, err
		}
	}
	for i := range cart.PRG {
		n, err := io.ReadFull(file, cart.PRG[i])
		if (err == io.EOF || err == io.ErrUnexpectedEOF) &&
			n == len(cart.PRG[i]) && int(header.NumCHRBanks) == 0 {
			break
		} else if err != nil {
			return nil, err
		}
	}
	if int(header.NumCHRBanks) > 0 {
		for i := range cart.CHR {
			_, err = io.ReadFull(file, cart.CHR[i])
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				return nil, err
			}
		}
	}
	mapperID := int((header.Control1 >> 4) | (header.Control2 & 0xf0))
	log.Printf("header control1: %b", header.Control1)
	log.Printf("header control2: %d", header.Control2)
	log.Printf("ROM: PRG-RPM: %d x 16KB  CHR-ROM %d x 8KB Mapper: %d", header.NumPRGBanks, header.NumCHRBanks, mapperID)
	// log.Info(mapperID)
	// For nestest.nes
	if mapperID == 171 {
		log.Debug("nestest.nes file")
		mapperID = 0
	}

	return cart, nil
}
