// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_VARIO
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	VarioFrameLength = 2 + 2 //Payload + Type + CRC
)

type VarioData struct {
	Speed int16 // cm/s (e.g. 1.5m/s sent as 150) Little-Endian
}

func UnmarshalVario(data []byte) (VarioData, error) {
	d := VarioData{}
	if len(data) != VarioFrameLength {
		return d, ErrFrameLength
	}
	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}
	//TODO check correct type?

	d.Speed = int16(binary.LittleEndian.Uint16(data[1:3]))
	return d, nil
}

func (d *VarioData) String() string {
	return fmt.Sprintf("Speed: %dcm/s", d.Speed)
}
