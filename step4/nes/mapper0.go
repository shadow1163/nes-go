package nes

import "fmt"

// Mapper0 implements the NROM mapper.
//
// http://wiki.nesdev.com/w/index.php/NROM
type Mapper0 struct {
	*Cartridge
	prgBank1 int
	prgBank2 int
}

// NewMapper0 create mapper 0
func NewMapper0(cart *Cartridge) *Mapper0 {
	m := &Mapper0{Cartridge: cart}

	numPRGBanks := len(cart.PRG)

	switch numPRGBanks {
	case 1:
		m.prgBank1 = 0
		m.prgBank2 = 0
	case 2:
		m.prgBank1 = 0
		m.prgBank2 = 1
	}

	return m
}

func (m *Mapper0) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return m.CHR[0][address]
	case address >= 0xC000:
		return m.PRG[m.prgBank2][address-0xC000]
	case address >= 0x8000:
		return m.PRG[m.prgBank1][address-0x8000]
	case address >= 0x6000:
		return m.SRAM[0][address-0x6000]
	default:
		log.Fatalf("Mapper 0 unhandle read address %x", address)
	}
	return 0
}

func (m *Mapper0) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		m.CHR[0][address] = value
	case address >= 0x8000:
		log.Warning(fmt.Sprintf("try to write address %x", address))
	case address >= 0x6000:
		m.SRAM[0][address-0x6000] = value
	default:
		log.Warning(fmt.Sprintf("Mapper 0 unhandle write address %x", address))
	}
}
