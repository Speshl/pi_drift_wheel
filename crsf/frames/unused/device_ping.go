// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_DEVICE_PING
package frames

import (
	"fmt"
)

const (
	DevicePingFrameLength = 2 + 2 //Payload + Type + CRC
)

type DevicePingData struct {
	Destination uint8
	Source      uint8
}

func UnmarshalDevicePing(data []byte) (DevicePingData, error) {
	d := DevicePingData{}
	if len(data) != DevicePingFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Destination = data[1]
	d.Source = data[2]

	//TODO CRC byte?
	return d, nil
}
