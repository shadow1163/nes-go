package nes

const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

// Joypad joypad struct
type Joypad struct {
	buttons [8]bool
	index   byte
	strobe  byte
}

//NewJoypad create a new Joypad
func NewJoypad() *Joypad {
	return &Joypad{}
}

// SetButtons set buttons
func (joy *Joypad) SetButtons(buttons [8]bool) {
	joy.buttons = buttons
}

// Read read Joypad
func (joy *Joypad) Read() byte {
	value := byte(0)
	if joy.index < 8 && joy.buttons[joy.index] {
		value = 1
	}
	joy.index++
	if joy.strobe&1 == 1 {
		joy.index = 0
	}
	return value
}

// Write write joypad
func (joy *Joypad) Write(value byte) {
	joy.strobe = value
	if joy.strobe&1 == 1 {
		joy.index = 0
	}
}
