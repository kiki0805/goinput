package goinput

import (
	"syscall"
	"unsafe"
)

const (
	// evSyn is used as markers to separate events. Events may be separated in time or in space, such as with the multitouch protocol.
	evSyn eventType = 0x00
	// evKey is used to describe state changes of keyboards, buttons, or other key-like devices.
	evKey eventType = 0x01
	// evRel is used to describe relative axis value changes, e.g. moving the mouse 5 units to the left.
	evRel eventType = 0x02
	// evAbs is used to describe absolute axis value changes, e.g. describing the coordinates of a touch on a touchscreen.
	evAbs eventType = 0x03
	// evMsc is used to describe miscellaneous input data that do not fit into other types.
	evMsc eventType = 0x04
	// evSw is used to describe binary state input switches.
	evSw eventType = 0x05
	// evLed is used to turn LEDs on devices on and off.
	evLed eventType = 0x11
	// evSnd is used to output sound to devices.
	evSnd eventType = 0x12
	// evRep is used for autorepeating devices.
	evRep eventType = 0x14
	// evFf is used to send force feedback commands to an input device.
	evFf eventType = 0x15
	// evPwr is a special type for power button and switch input.
	evPwr eventType = 0x16
	// evFfStatus is used to receive force feedback device status.
	evFfStatus eventType = 0x17
)

// eventType are groupings of codes under a logical input construct.
// Each type has a set of applicable codes to be used in generating events.
// See the Ev section for details on valid codes for each type
type eventType uint16

// eventsize is size of structure of inputEvent
var eventsize = int(unsafe.Sizeof(inputEvent{}))

// inputEvent is the keyboard event structure itself
type inputEvent struct {
	Time  syscall.Timeval
	Type  eventType
	Code  uint16
	Value int32
}

// KeyString returns representation of pressed key as string
// eg enter, space, a, b, c...
func (i *inputEvent) String() string {
	return keyCodeMap[i.Code]
}

// KeyPress is the value when we press the key on keyboard
func (i *inputEvent) KeyPress() bool {
	return i.Value == 1
}

// KeyRelease is the value when we release the key on keyboard
func (i *inputEvent) KeyRelease() bool {
	return i.Value == 0
}

// Button is mouse button with Name known
type Button struct {
	code int
	Name string
}

// Key is key on keyboard with Name known
type Key struct {
	code uint16
	Name string
}

// MouseClick is click event with Press indicator and involved Button
type MouseClick struct {
	Button
	Press bool
}

// MouseMove is mouse movement event with current location: (X, Y)
type MouseMove struct {
	X int
	Y int
}

// MouseScroll is mouse scroll event
type MouseScroll struct {
	Dx int
	Dy int
}

// KeyPress is key pressed event on keyboard
type KeyPress struct {
	Key
}

// KeyRelease is key released event on keyboard
type KeyRelease struct {
	Key
}
