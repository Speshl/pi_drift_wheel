package sbus

import (
	"fmt"
)

const (
	frameLength int    = 25
	startByte   byte   = 0x0f
	endByte     byte   = 0x00
	mask        uint16 = 0x07ff // The maximum 11-bit channel value

	MaxChannels int = 16

	//MinValue int = 272
	MinValue int = 172
	MidValue int = 992
	MaxValue int = 1811
	//MaxValue int = 1712
)

// Channels is the ordered list of servo channel values with 16 channels
type Channels [MaxChannels]uint16

// Flags stores SBUS flags
type Flags struct {
	Ch17      bool
	Ch18      bool
	Framelost bool
	Failsafe  bool
}

// Frame is an SBUS data frame with 16 proportional channels
type Frame struct {
	Ch    Channels
	Flags Flags
}

func NewFrame() Frame {
	frameChannels := [MaxChannels]uint16{}
	for i := range frameChannels {
		frameChannels[i] = uint16(MidValue)
	}
	return Frame{
		Ch: frameChannels,
	}
}

// Shows channels as their bits
func (c Channels) String() {
	fmt.Printf("% 8b", c)
}

// Shows flags as their bits
func (f Flags) String() {
	fmt.Printf("% 8b", f.marshal())
}

func (f Flags) marshal() (fbyte byte) {
	if f.Ch17 {
		fbyte |= 0x80
	}
	if f.Ch18 {
		fbyte |= 0x40
	}
	if f.Framelost {
		fbyte |= 0x20
	}
	if f.Failsafe {
		fbyte |= 0x10
	}
	return
}

// Bit shifts taken from: https://github.com/johnelliott/go-sbus
// Marshal serializes a Frame to bytes
func (f Frame) Marshal() []byte {
	return []byte{
		startByte,
		byte(f.Ch[0] & mask),
		byte((f.Ch[0]&mask)>>8 | (f.Ch[1]&mask)<<3),
		byte((f.Ch[1]&mask)>>5 | (f.Ch[2]&mask)<<6),
		byte((f.Ch[2] & mask) >> 2),
		byte((f.Ch[2]&mask)>>10 | (f.Ch[3]&mask)<<1),
		byte((f.Ch[3]&mask)>>7 | (f.Ch[4]&mask)<<4),
		byte((f.Ch[4]&mask)>>4 | (f.Ch[5]&mask)<<7),
		byte((f.Ch[5] & mask) >> 1),
		byte((f.Ch[5]&mask)>>9 | (f.Ch[6]&mask)<<2),
		byte((f.Ch[6]&mask)>>6 | (f.Ch[7]&mask)<<5),
		byte((f.Ch[7] & mask) >> 3),
		byte((f.Ch[8] & mask)),
		byte((f.Ch[8]&mask)>>8 | (f.Ch[9]&mask)<<3),
		byte((f.Ch[9]&mask)>>5 | (f.Ch[10]&mask)<<6),
		byte((f.Ch[10] & mask) >> 2),
		byte((f.Ch[10]&mask)>>10 | (f.Ch[11]&mask)<<1),
		byte((f.Ch[11]&mask)>>7 | (f.Ch[12]&mask)<<4),
		byte((f.Ch[12]&mask)>>4 | (f.Ch[13]&mask)<<7),
		byte((f.Ch[13] & mask) >> 1),
		byte((f.Ch[13]&mask)>>9 | (f.Ch[14]&mask)<<2),
		byte((f.Ch[14]&mask)>>6 | (f.Ch[15]&mask)<<5),
		byte((f.Ch[15] & mask) >> 3),
		f.Flags.marshal(),
		endByte,
	}
}

// UnmarshalFrame tries to create a Frame from a byte array
func UnmarshalFrame(data []byte) (f Frame, err error) {
	if len(data) != frameLength {
		err = fmt.Errorf("incorrect frame size")
		return
	}
	if data[0] != startByte {
		err = fmt.Errorf("error parsing frame: incorrect start byte %v", data[0])
		return
	}
	if data[frameLength-1] != endByte {
		err = fmt.Errorf("error parsing frame: incorrect end byte %v", data[frameLength-1])
		return
	}

	f.Ch[0] = ((uint16(data[1]) | uint16(data[2])<<8) & mask)
	f.Ch[1] = ((uint16(data[2])>>3 | uint16(data[3])<<5) & mask)
	f.Ch[2] = ((uint16(data[3])>>6 | uint16(data[4])<<2 | uint16(data[5])<<10) & mask)
	f.Ch[3] = ((uint16(data[5])>>1 | uint16(data[6])<<7) & mask)
	f.Ch[4] = ((uint16(data[6])>>4 | uint16(data[7])<<4) & mask)
	f.Ch[5] = ((uint16(data[7])>>7 | uint16(data[8])<<1 | uint16(data[9])<<9) & mask)
	f.Ch[6] = ((uint16(data[9])>>2 | uint16(data[10])<<6) & mask)
	f.Ch[7] = ((uint16(data[10])>>5 | uint16(data[11])<<3) & mask)
	f.Ch[8] = ((uint16(data[12]) | uint16(data[13])<<8) & mask)
	f.Ch[9] = ((uint16(data[13])>>3 | uint16(data[14])<<5) & mask)
	f.Ch[10] = ((uint16(data[14])>>6 | uint16(data[15])<<2 | uint16(data[16])<<10) & mask)
	f.Ch[11] = ((uint16(data[16])>>1 | uint16(data[17])<<7) & mask)
	f.Ch[12] = ((uint16(data[17])>>4 | uint16(data[18])<<4) & mask)
	f.Ch[13] = ((uint16(data[18])>>7 | uint16(data[19])<<1 | uint16(data[20])<<9) & mask)
	f.Ch[14] = ((uint16(data[20])>>2 | uint16(data[21])<<6) & mask)
	f.Ch[15] = ((uint16(data[21])>>5 | uint16(data[22])<<3) & mask)

	f.Flags.Failsafe = (data[frameLength-2] & 0x10) != 0
	f.Flags.Framelost = (data[frameLength-2] & 0x20) != 0
	f.Flags.Ch18 = (data[frameLength-2] & 0x40) != 0
	f.Flags.Ch17 = (data[frameLength-2] & 0x80) != 0

	return
}
