package nes

import (
	"fmt"
)

// Mapper mapper interface
type Mapper interface {
	Read(address uint16) byte
	Write(address uint16, value byte)
}

// NewMapper create a mapper
func NewMapper(id int, cart *Cartridge) (Mapper, error) {
	var mapper Mapper
	switch id {
	case 0:
		mapper = NewMapper0(cart)
	default:
		return nil, fmt.Errorf("mapper ID %d not implemented", id)
	}
	return mapper, nil
}
