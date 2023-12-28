// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_LINK_RX_ID
package frames

import (
	"fmt"
)

const (
	LinkRxFrameLength = 4 + 2 //Payload + Type + CRC
)

type LinkRxData struct {
	RssiPercent int8
	Unknown1    uint8
	Unknown2    uint8
	PowerIndex  int8
}

func UnmarshalLinkRx(data []byte) (LinkRxData, error) {
	d := LinkRxData{}
	if len(data) != LinkRxFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.RssiPercent = int8(data[1])
	d.Unknown1 = uint8(data[2])
	d.Unknown2 = uint8(data[3])
	d.PowerIndex = int8(data[4])

	//TODO CRC byte?
	return d, nil
}

func (d *LinkRxData) String() string {
	return fmt.Sprintf("RssiPercent: %d%% Unknown1: %d Unknown2: %d PowerIndex: %d", d.RssiPercent, d.Unknown1, d.Unknown2, d.PowerIndex)
}
