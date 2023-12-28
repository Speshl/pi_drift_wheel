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
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Pitch = int16(binary.LittleEndian.Uint16(data[1:3]))
	d.Roll = int16(binary.LittleEndian.Uint16(data[3:5]))
	d.Yaw = int16(binary.LittleEndian.Uint16(data[5:7]))

	//TODO CRC byte?
	return d, nil
}

func (d *AttitudeData) String() string {
	pitch := (float32(d.Pitch) / 10000) * (180 / 3.14159)
	roll := (float32(d.Roll) / 10000) * (180 / 3.14159)
	yaw := (float32(d.Yaw) / 10000) * (180 / 3.14159)

	return fmt.Sprintf("Pitch: %.2f Roll: %.2f Yaw: %.2f", pitch, roll, yaw)
}
