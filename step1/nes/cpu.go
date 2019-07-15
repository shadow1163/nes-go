package nes

// CPU nes cpu struct
type CPU struct {
	Cart *Cartridge
	RAM  [2048]byte
}

// NewCPU create a nes cpu
func NewCPU(cart *Cartridge) *CPU {
	cpu := CPU{Cart: cart}
	// cpu.Reset()
	return &cpu
}

func (cpu *CPU) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return cpu.RAM[address%0x0800]
	case address < 0x4000:
		log.Warning("Not Imp")
	case address == 0x4014:
		log.Warning("Not Imp")
	case address == 0x4015:
		log.Warning("Not Imp")
	case address == 0x4016:
		log.Warning("Not Imp")
	case address == 0x4017:
		log.Warning("Not Imp")
	case address < 0x6000:
		// TODO: I/O registers
	case address >= 0x6000:
		return cpu.Cart.Mapper.Read(address)
	default:
		log.Fatalf("unhandled cpu memory read at address: 0x%04X", address)
	}
	return 0
}

func (cpu *CPU) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		cpu.RAM[address%0x0800] = value
	case address < 0x4000:
		log.Warning("Not Imp")
	case address == 0x4014:
		log.Warning("Not Imp")
	case address == 0x4015:
		log.Warning("Not Imp")
	case address == 0x4016:
		log.Warning("Not Imp")
	case address == 0x4017:
		log.Warning("Not Imp")
	case address < 0x6000:
		// TODO: I/O registers
	case address >= 0x6000:
		cpu.Cart.Mapper.Write(address, value)
	default:
		log.Fatalf("unhandled cpu memory write at address: 0x%04X", address)
	}
}

// Read16 cpu read 2 bytes
func (cpu *CPU) Read16(address uint16) uint16 {
	highAddress := address + 1

	// log.Debug(fmt.Sprintf("%02x", cpu.Read(address)))
	// log.Debug(fmt.Sprintf("%02x", cpu.Read(highAddress)))

	return uint16(cpu.Read(address)) | uint16(cpu.Read(highAddress))<<8
}
