package frames

import "fmt"

var (
	ErrFrameLength = fmt.Errorf("incorrect frame length")
	ErrInvalidCRC8 = fmt.Errorf("frame failed crc8 validation")
)

func Crc8DVB_S2(crc, a uint8) uint8 {
	crc = crc ^ a
	for ii := 0; ii < 8; ii++ {
		if crc&0x80 != 0 {
			crc = (crc << 1) ^ 0xD5
		} else {
			crc = crc << 1
		}
	}
	return crc & 0xFF
}

func GenerateCrc8Value(data []uint8) uint8 {
	crc := uint8(0)
	for _, value := range data {
		crc = Crc8DVB_S2(crc, value)
	}
	return crc
}

func ValidateFrame(frame []uint8) bool {
	frameSize := len(frame)
	crc := GenerateCrc8Value(frame[0 : frameSize-1])
	return crc == frame[frameSize-1]
}
