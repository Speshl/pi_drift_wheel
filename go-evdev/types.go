package evdev

import (
	"fmt"
	"syscall"
)

// EvType is EV_KEY, EV_SW, EV_LED, EV_SND, ...
type EvType uint16

// EvCode describes codes within a type (eg. KEY_A, KEY_B, ...)
type EvCode uint16

// EvProp describes device properties (eg. INPUT_PROP_ACCELEROMETER, INPUT_PROP_BUTTONPAD, ...)
type EvProp uint16

// StateMap describes the current state of codes within a type, as booleans.
type StateMap map[EvCode]bool

// InputEvent describes an event that is generated by an InputDevice
type InputEvent struct {
	Time  syscall.Timeval // time in seconds since epoch at which event occurred
	Type  EvType          // event type - one of ecodes.EV_*
	Code  EvCode          // event code related to the event type
	Value int32           // event value related to the event type
}

func (e *InputEvent) TypeName() string {
	return TypeName(e.Type)
}

func (e *InputEvent) CodeName() string {
	return CodeName(e.Type, e.Code)
}

func (e *InputEvent) String() string {
	return fmt.Sprintf(
		"type: 0x%02x [%s], code: 0x%02x [%s], value: %d",
		e.Type, e.TypeName(), e.Code, e.CodeName(), e.Value,
	)
}

// InputID ...
type InputID struct {
	BusType uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// AbsInfo describes details on ABS input types
type AbsInfo struct {
	Value      int32
	Minimum    int32
	Maximum    int32
	Fuzz       int32
	Flat       int32
	Resolution int32
}

// InputKeymapEntry is used to retrieve and modify keymap data
type InputKeymapEntry struct {
	Flags    uint8
	Len      uint8
	Index    uint16
	KeyCode  uint32
	ScanCode [32]uint8
}

// InputMask ...
type InputMask struct {
	Type      uint32
	CodesSize uint32
	CodesPtr  uint64
}

// UinputUserDevice is used when creating or cloning a device
type UinputUserDevice struct {
	Name       [uinputMaxNameSize]byte
	ID         InputID
	EffectsMax uint32
	Absmax     [absSize]int32
	Absmin     [absSize]int32
	Absfuzz    [absSize]int32
	Absflat    [absSize]int32
}

// Used to build up force feedback effects
type Envelope struct { // 8 bytes
	AttackLength uint16
	AttackLevel  uint16
	FadeLength   uint16
	FadeLevel    uint16
}

type Constant struct { //10 bytes, padded to 24
	Level    int16
	Envelope Envelope

	unused [14]byte
}

type Rumble struct {
	Strong uint16
	Weak   uint16
}

type Periodic struct {
	Waveform     uint16
	Period       uint16
	Magnitude    int16
	Offset       int16
	Phase        uint16
	Envelope     Envelope
	CustomLength uint32
	CustomDate   *int16
}

type Condition struct {
	RightSaturation uint16
	LeftSaturation  uint16
	RightCoeff      int16
	LeftCoeff       int16
	Deadband        uint16
	Center          uint16
}

type Ramp struct {
	Start    int16
	End      int16
	Envelope Envelope
}

type Replay struct { //4 bytes
	Length uint16
	Delay  uint16
}

type Trigger struct { //4 bytes
	Button   uint16
	Interval uint16
}

type EffectType struct {
	Constant Constant //10 bytes, padded to 24
	// Ramp      Ramp //8 bytes
	// Periodic  Periodic //20 bytes
	// Condition [2]Condition //one for each axis 12 (24)
	// Rumble    Rumble //4 bytes
}

type Effect struct { //38 bytes
	Type       uint16
	Id         int16
	Direction  uint16
	Trigger    Trigger
	Replay     Replay
	EffectType EffectType
}
