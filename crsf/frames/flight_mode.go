// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_FLIGHT_MODE
package frames

import (
	"fmt"
)

const (
	FlightModeFrameLength = 14 + 2 //Payload + Type + CRC
)

type FlightModeData struct {
	FlightMode string //max length 13
}

func UnmarshalFlightMode(data []byte) (FlightModeData, error) {
	d := FlightModeData{}
	if len(data) > 3 {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	for i := range data[1:15] {
		if i == 0x00 { //null terminator for string
			break
		}
		d.FlightMode += string(data[i])
	}

	//TODO CRC byte?
	return d, nil
}

func (d *FlightModeData) String() string {
	return fmt.Sprintf("FlightMode: %s", d.FlightMode)
}
