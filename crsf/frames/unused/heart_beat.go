// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_HEARTBEAT
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	HeartBeatFrameLength = 2 + 2 //Payload + Type + CRC
)

type HeartBeatData struct {
	Origin uint16 //Origin device address big-endian
}

func UnmarshalHeartBeat(data []byte) (HeartBeatData, error) {
	d := HeartBeatData{}
	if len(data) != HeartBeatFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Origin = binary.BigEndian.Uint16(data[1:3])

	//TODO CRC byte?
	return d, nil
}
