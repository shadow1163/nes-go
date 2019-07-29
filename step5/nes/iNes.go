package nes

type iNESHeader struct {
	Magic        [3]byte
	Format       byte
	NumPRGBanks  byte
	NumCHRBanks  byte
	Control1     byte
	Control2     byte
	NumSRAMBanks byte
	_            [7]byte
}
