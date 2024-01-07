// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_ATTITUDE
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	AttitudeFrameLength = 6 + 2 //Payload + Type + CRC
)

// All values must be in the +/-180 degree +/-PI radian range
type AttitudeData struct {
	Pitch int16 //angle in radians * 10000
	Roll  int16 //angle in radians * 10000
	Yaw   int16 //angle in radians * 10000
}

func UnmarshalAttitude(data []byte) (AttitudeData, error) {
	d := AttitudeData{}
	if len(data) != AttitudeFrameLength {
		return d, ErrFrameLength
	}

	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}
	//TODO check correct type?

	d.Pitch = int16(binary.BigEndian.Uint16(data[1:3]))
	d.Roll = int16(binary.BigEndian.Uint16(data[3:5]))
	d.Yaw = int16(binary.BigEndian.Uint16(data[5:7]))

	return d, nil
}

func (d *AttitudeData) String() string {
	pitch := getAsDegree(d.Pitch)
	roll := getAsDegree(d.Roll)
	yaw := getAsDegree(d.Yaw)

	return fmt.Sprintf("Pitch: %.2f Roll: %.2f Yaw: %.2f", pitch, roll, yaw)
}

func (d *AttitudeData) PitchDegree() float64 {
	return getAsDegree(d.Pitch)
}

func (d *AttitudeData) RollDegree() float64 {
	return getAsDegree(d.Roll)
}

func (d *AttitudeData) YawDegree() float64 {
	return getAsDegree(d.Yaw)
}

func getAsDegree(value int16) float64 {
	return (float64(value) / 10000) * (180 / 3.14159)
}
