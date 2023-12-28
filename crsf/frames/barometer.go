// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_BARO_ALTITUDE
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	BarometerFrameLength = 4 + 2 //Payload + Type + CRC
)

type BarometerData struct {
	/*
		If high bit IS NOT set, value is in decimeters + 10000  values between(-1000.0m - 2276.7m)
		If high bit IS set, value is in meters with values between (0m-32767m)
	*/
	Altitude uint16

	Speed int16 // cm/s (1.5m/s is 150)
}

func UnmarshalBarometer(data []byte) (BarometerData, error) {
	d := BarometerData{}
	if len(data) != BarometerFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Altitude = binary.LittleEndian.Uint16(data[1:3])
	d.Speed = int16(binary.LittleEndian.Uint16(data[3:5]))

	//TODO CRC byte?
	return d, nil
}

func (d *BarometerData) String() string {

	altitude := float32(0)
	if d.Altitude > 32768 { //highest value a uint16 can have before high bit is set
		//high bit IS set so value is in meters
		altitude = float32(d.Altitude) - 32768
	} else {
		//high bit IS NOT set so value is in decimeters
		altitude = (float32(d.Altitude) - 10000) / 10
	}

	speed := d.Speed

	return fmt.Sprintf("Altitude: %.1fm Speed: %dcm/s", altitude, speed)
}
