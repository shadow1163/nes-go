package nes

import (
	"fmt"
	"image"
	"image/color"
)

// PPU nes ppu struct
type PPU struct {
	CPU  *CPU
	Cart *Cartridge

	// Screen image, 256x240px.
	img *image.RGBA

	// NES fixed 64 colour palette.
	palette [64]color.RGBA

	// Scanline (0-261).
	Scanline int

	// Tick (0-340).
	Tick int

	// Frame counter.
	Frame uint64

	// Total number of cycles executed.
	numCycles uint64

	// RAM.
	ram [16384]byte

	// Sprite RAM.
	sprRAM [256]byte

	// PPU Control Register 1 ($2000).
	spriteTableAddress     uint16
	backgroundTableAddress uint16
	flagIncrementBy32      bool
	flagLargeSprites       bool
	flagNMIOnVBlank        bool

	// PPU Control Register 2 ($2001).
	flagColourMode     bool
	flagClipBackground bool
	flagClipSprites    bool
	flagShowBackground bool
	flagShowSprites    bool
	flagRedEmphasis    bool
	flagGreenEmphasis  bool
	flagBlueEmphasis   bool

	// Misc flags.
	flagVRAMWritesIgnored  bool
	flagScanlineSpritesMax bool
	flagSprite0Hit         bool
	flagVBlankOutstanding  bool

	// Internal registers.
	v uint16 // Current VRAM address (15 bits).
	t uint16 // Temporary VRAM address. (15 bits).
	x byte   // Fine X scroll (3 bits).
	w byte   // First or second write toggle (0=first, 1=second).

	// The next 16 pixels of background.
	bgPixels [16]*color.RGBA

	// Sprite IO address.
	sprIOAddress byte

	// PPUDATA read buffer.
	readBuffer byte
}

// NewPPU get a ppu instance
func NewPPU(cart *Cartridge, cpu *CPU) *PPU {
	ppu := &PPU{
		Cart:     cart,
		CPU:      cpu,
		Scanline: 241,
		Tick:     0}

	ppu.setupPalette()
	ppu.flagShowBackground = true

	return ppu
}

// Step ppu exeute a step
func (ppu *PPU) Step() *image.RGBA {
	ppu.tick()

	var outputImage *image.RGBA

	isRendering := ppu.flagShowBackground
	isVisible := ppu.Scanline <= 239
	isVBlanline := ppu.Scanline == 241
	isPrerender := ppu.Scanline == 261
	isDrawing := isRendering && isVisible &&
		((ppu.Tick >= 1 && ppu.Tick <= 256) || (ppu.Tick >= 321 && ppu.Tick <= 336))

	isFetching := isRendering && (isVisible || isPrerender) &&
		((ppu.Tick >= 1 && ppu.Tick <= 256) || (ppu.Tick >= 321 && ppu.Tick <= 336)) &&
		ppu.Tick%8 == 0
	if isDrawing {
		ppu.drawPixel()
	}

	if isFetching {
		// ppu.loadTile()
	}

	if isVBlanline && ppu.Tick == 1 {
		ppu.flagVBlankOutstanding = true
		if ppu.flagNMIOnVBlank {
			ppu.CPU.NMI()
		}
		outputImage = ppu.img
	} else if isPrerender && ppu.Tick == 1 {
		// Clear flags.
		ppu.flagVBlankOutstanding = false
		// ppu.flagScanlineSpritesMax = false
		// ppu.flagSprite0Hit = false
	}
	return outputImage
}

// ReadRegister read ppu register
func (ppu *PPU) ReadRegister(address uint16) byte {
	switch address {
	case 0x2002:
		return ppu.readStatus()
	case 0x2004:
		return ppu.readOAMData()
	case 0x2007:
		return ppu.readData()
	}
	return 0
}

func (ppu *PPU) String() string {
	return fmt.Sprintf("PPU[Scanline=%d, Tick=%d, CX=%d, FX=%d, CY=%d, FY=%d, NT=%d, R=%v]",
		ppu.Scanline,
		ppu.Tick,
		ppu.v&0x001F,
		ppu.x,
		(ppu.v&0x03E0)>>5,
		(ppu.v&0x7000)>>12,
		(ppu.v&0x0800)>>11,
		ppu.flagShowBackground || ppu.flagShowSprites)
}

//DoVBlank do vblank
func (ppu *PPU) DoVBlank() {
	ppu.flagVBlankOutstanding = true
	if ppu.flagNMIOnVBlank {
		ppu.CPU.NMI()
	}
}

// GetPixel Get x, y pixel
func (ppu *PPU) GetPixel(x, y int) color.RGBA {
	id := x>>3 + (y>>3)*32
	name := ppu.read(0x2000 + uint16(id))
	// log.Debug(fmt.Sprintf("%d, %d -> 0x%02x", x, y, name))
	offset := y & 0x7
	p0 := ppu.read(uint16(name)<<4 + uint16(offset))
	p1 := ppu.read(uint16(name)<<4 + 8 + uint16(offset))

	// log.Debug(fmt.Sprintf("%d, %d -> p0: 0x%02x", x, y, p0))
	// log.Debug(fmt.Sprintf("p1: 0x%02x", p1))

	shift := (^x) & 0x7
	mask := 1 << uint(shift)

	low := (uint(p0) & uint(mask) >> uint(shift)) | ((uint(p1) & uint(mask)) >> uint(shift) << 1)

	aid := (x >> 5) + (y>>5)*8
	// log.Debug(fmt.Sprintf("aid: 0x%02x", aid))
	attr := ppu.read(0x23c0 + uint16(aid))
	// log.Debug(fmt.Sprintf("attr: 0x%02x", attr))
	aoffset := ((x & 0x10) >> 3) | ((y & 0x10) >> 2)
	// log.Debug(fmt.Sprintf("aoffset: 0x%02x", aoffset))
	high := (attr & (3 << uint(aoffset))) >> uint(aoffset) << 2

	index := uint(low) | uint(high)
	// log.Debug(fmt.Sprintf("index: 0x%02x", index))
	paletteIndex := ppu.read(0x3F00+uint16(index)) & 0x3F
	// log.Debug(fmt.Sprintf("paletteIndex: 0x%02x", paletteIndex))
	return ppu.palette[paletteIndex]
}

// WriteRegister write to ppu register
func (ppu *PPU) WriteRegister(address uint16, value byte) {
	switch address {
	case 0x2000:
		ppu.writeControl(value)
	case 0x2001:
		ppu.writeMask(value)
	case 0x2003:
		ppu.writeOAMAddress(value)
	case 0x2004:
		ppu.writeOAMData(value)
	case 0x2005:
		ppu.writeScroll(value)
	case 0x2006:
		ppu.writeAddress(value)
	case 0x2007:
		ppu.writeData(value)
	default:
		log.Error(fmt.Sprintf("Unknown write @ %x", address))
	}
}

func (ppu *PPU) readStatus() byte {
	var result byte
	if ppu.flagScanlineSpritesMax {
		result |= 0x20
	}

	if ppu.flagSprite0Hit {
		result |= 0x40
	}

	if ppu.flagVBlankOutstanding {
		result |= 0x80
		ppu.flagVBlankOutstanding = false
	}

	ppu.w = 0
	return result
}

func (ppu *PPU) readOAMData() byte {
	return ppu.sprRAM[ppu.sprIOAddress]
}

func (ppu *PPU) readData() byte {
	previousValue := ppu.readBuffer
	ppu.readBuffer = ppu.read(ppu.v)

	var result byte

	if ppu.v&0x3FFF <= 0x3EFF {
		result = previousValue
	} else {
		result = ppu.readBuffer
	}

	if ppu.flagIncrementBy32 {
		ppu.v += 32
	} else {
		ppu.v++
	}
	return result
}

func (ppu *PPU) writeControl(value byte) {
	// t: ...BA.. ........ = d: ......BA
	ppu.t = ppu.t&0x73FF | (uint16(value)&0x3)<<10

	ppu.flagIncrementBy32 = value&0x4 != 0

	if value&0x8 == 0 {
		ppu.spriteTableAddress = 0x0000
	} else {
		ppu.spriteTableAddress = 0x1000
	}

	if value&0x10 == 0 {
		ppu.backgroundTableAddress = 0x0000
	} else {
		ppu.backgroundTableAddress = 0x1000
	}

	ppu.flagLargeSprites = value&0x20 != 0
	ppu.flagNMIOnVBlank = value&0x80 != 0
}

func (ppu *PPU) writeMask(value byte) {
	ppu.flagColourMode = value&0x1 == 0

	ppu.flagClipBackground = value&0x2 == 0
	ppu.flagClipSprites = value&0x4 == 0

	ppu.flagShowBackground = value&0x8 != 0
	ppu.flagShowSprites = value&0x10 != 0

	ppu.flagRedEmphasis = value&0x20 != 0
	ppu.flagGreenEmphasis = value&0x40 != 0
	ppu.flagBlueEmphasis = value&0x80 != 0
}

func (ppu *PPU) writeOAMAddress(value byte) {
	ppu.sprIOAddress = value
}

func (ppu *PPU) writeOAMData(value byte) {
	ppu.sprRAM[ppu.sprIOAddress] = value
	ppu.sprIOAddress++
}

func (ppu *PPU) writeScroll(value byte) {
	if ppu.w == 0 {
		// t: ....... ...HGFED = d: HGFED...
		// x:              CBA = d: .....CBA
		// w:                  = 1
		ppu.t = (ppu.t & 0xFFE0) | ((uint16(value) & 0xF8) >> 3)
		ppu.x = value & 0x7
		ppu.w = 1
	} else {
		// t: CBA..HG FED..... = d: HGFEDCBA
		// w:                  = 0
		ppu.t = (ppu.t & 0x0C1F) | ((uint16(value) & 0x7) << 12) |
			((uint16(value) & 0xF8) << 2)
		ppu.w = 0
	}
}

func (ppu *PPU) writeAddress(value byte) {
	if ppu.w == 0 {
		// t: .FEDCBA ........ = d: ..FEDCBA
		// t: X...... ........ = 0
		// w:                  = 1
		ppu.t = (ppu.t & 0x00FF) | ((uint16(value) & 0x3F) << 8)
		ppu.t &= 0x7FFF
		ppu.w = 1
	} else {
		// t: ....... HGFEDCBA = d: HGFEDCBA
		// v                   = t
		// w:                  = 0
		ppu.t = (ppu.t & 0xFF00) | uint16(value)
		ppu.v = ppu.t
		ppu.w = 0
	}
}

func (ppu *PPU) writeData(value byte) {
	ppu.write(ppu.v, value)
	if ppu.flagIncrementBy32 {
		ppu.v += 32
	} else {
		ppu.v++
	}
}

func (ppu *PPU) drawPixel() {

	// ppu.img.Set()
}

func (ppu *PPU) read(address uint16) byte {
	address = ppu.mapAddress(address)

	switch {
	case address < 0x2000:
		// log.Debug(fmt.Sprintf("read from address 0x%04x", address))
		return ppu.Cart.Mapper.Read(address)
	default:
		return ppu.ram[address]
	}
}

func (ppu *PPU) write(address uint16, value byte) {
	address = ppu.mapAddress(address)

	switch {
	case address < 0x2000:
		ppu.Cart.Mapper.Write(address, value)
	default:
		ppu.ram[address] = value
	}
}

func (ppu *PPU) mapAddress(address uint16) uint16 {
	address &= 0x3FFF

	// Sprite palette mirroring.
	if address == 0x3F10 ||
		address == 0x3F14 ||
		address == 0x3F18 ||
		address == 0x3F1C {
		address -= 0x10
	} else if address >= 0x2000 && address <= 0x2FFF {
		// Nametable mirroring.
		mirror := ppu.Cart.Mirror

		if mirror == horizontal {
			if address >= 0x2400 && address < 0x2800 {
				address -= 0x400
			} else if address >= 0x2C00 && address < 0x3000 {
				address -= 0x400
			}
		} else if mirror == vertical {
			if address >= 0x2800 && address < 0x2C00 {
				address -= 0x800
			} else if address >= 0x2C00 && address < 0x3000 {
				address -= 0x800
			}
		} else if mirror == singleLow {
			address = 0x2000 | (address & 0x3FF)
		} else if mirror == singleHigh {
			address = 0x2400 | (address & 0x3FF)
			//log.Printf("address %x mirrored to %x\n", orig, address)
		} else if mirror == fourScreen {
			// No mirroring.
		}
	}

	return address
}

func (ppu *PPU) tick() {
	ppu.Tick++

	isOddFrame := ppu.Frame&0x1 != 0

	if ppu.Scanline == 261 && (ppu.Tick == 341 || (ppu.Tick == 340 && isOddFrame)) {
		ppu.Scanline = 0
		ppu.Tick = 0
		ppu.Frame++
	} else if ppu.Tick == 341 {
		ppu.Scanline++
		ppu.Tick = 0
	}
}

func (ppu *PPU) setupPalette() {
	ppu.palette = [64]color.RGBA{
		/* 0x00 */ {0x75, 0x75, 0x75, 0xFF},
		/* 0x01 */ {0x27, 0x1B, 0x8F, 0xFF},
		/* 0x02 */ {0x00, 0x00, 0xAB, 0xFF},
		/* 0x03 */ {0x47, 0x00, 0x9F, 0xFF},
		/* 0x04 */ {0x8F, 0x00, 0x77, 0xFF},
		/* 0x05 */ {0xAB, 0x00, 0x13, 0xFF},
		/* 0x06 */ {0xA7, 0x00, 0x00, 0xFF},
		/* 0x07 */ {0x7F, 0x0B, 0x00, 0xFF},
		/* 0x08 */ {0x43, 0x2F, 0x00, 0xFF},
		/* 0x09 */ {0x00, 0x47, 0x00, 0xFF},
		/* 0x0A */ {0x00, 0x51, 0x00, 0xFF},
		/* 0x0B */ {0x00, 0x3F, 0x17, 0xFF},
		/* 0x0C */ {0x1B, 0x3F, 0x5F, 0xFF},
		/* 0x0D */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x0E */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x0F */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x10 */ {0xBC, 0xBC, 0xBC, 0xFF},
		/* 0x11 */ {0x00, 0x73, 0xEF, 0xFF},
		/* 0x12 */ {0x23, 0x3B, 0xEF, 0xFF},
		/* 0x13 */ {0x83, 0x00, 0xF3, 0xFF},
		/* 0x14 */ {0xBF, 0x00, 0xBF, 0xFF},
		/* 0x15 */ {0xE7, 0x00, 0x5B, 0xFF},
		/* 0x16 */ {0xDB, 0x2B, 0x00, 0xFF},
		/* 0x17 */ {0xCB, 0x4F, 0x0F, 0xFF},
		/* 0x18 */ {0x8B, 0x73, 0x00, 0xFF},
		/* 0x19 */ {0x00, 0x97, 0x00, 0xFF},
		/* 0x1A */ {0x00, 0xAB, 0x00, 0xFF},
		/* 0x1B */ {0x00, 0x93, 0x3B, 0xFF},
		/* 0x1C */ {0x00, 0x83, 0x8B, 0xFF},
		/* 0x1D */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x1E */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x1F */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x20 */ {0xFF, 0xFF, 0xFF, 0xFF},
		/* 0x21 */ {0x3F, 0xBF, 0xFF, 0xFF},
		/* 0x22 */ {0x5F, 0x97, 0xFF, 0xFF},
		/* 0x23 */ {0xA7, 0x8B, 0xFD, 0xFF},
		/* 0x24 */ {0xF7, 0x7B, 0xFF, 0xFF},
		/* 0x25 */ {0xFF, 0x77, 0xB7, 0xFF},
		/* 0x26 */ {0xFF, 0x77, 0x63, 0xFF},
		/* 0x27 */ {0xFF, 0x9B, 0x3B, 0xFF},
		/* 0x28 */ {0xF3, 0xBF, 0x3F, 0xFF},
		/* 0x29 */ {0x83, 0xD3, 0x13, 0xFF},
		/* 0x2A */ {0x4F, 0xDF, 0x4B, 0xFF},
		/* 0x2B */ {0x58, 0xF8, 0x98, 0xFF},
		/* 0x2C */ {0x00, 0xEB, 0xDB, 0xFF},
		/* 0x2D */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x2E */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x2F */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x30 */ {0xFF, 0xFF, 0xFF, 0xFF},
		/* 0x31 */ {0xAB, 0xE7, 0xFF, 0xFF},
		/* 0x32 */ {0xC7, 0xD7, 0xFF, 0xFF},
		/* 0x33 */ {0xD7, 0xCB, 0xFF, 0xFF},
		/* 0x34 */ {0xFF, 0xC7, 0xFF, 0xFF},
		/* 0x35 */ {0xFF, 0xC7, 0xDB, 0xFF},
		/* 0x36 */ {0xFF, 0xBF, 0xB3, 0xFF},
		/* 0x37 */ {0xFF, 0xDB, 0xAB, 0xFF},
		/* 0x38 */ {0xFF, 0xE7, 0xA3, 0xFF},
		/* 0x39 */ {0xE3, 0xFF, 0xA3, 0xFF},
		/* 0x3A */ {0xAB, 0xF3, 0xBF, 0xFF},
		/* 0x3B */ {0xB3, 0xFF, 0xCF, 0xFF},
		/* 0x3C */ {0x9F, 0xFF, 0xF3, 0xFF},
		/* 0x3D */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x3E */ {0x00, 0x00, 0x00, 0xFF},
		/* 0x3F */ {0x00, 0x00, 0x00, 0xFF},
	}
}
