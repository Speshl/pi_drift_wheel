// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_LINK_TX_ID
package frames

import (
	"fmt"
)

const (
	LinkTxFrameLength = 5 + 2 //Payload + Type + CRC
)

type LinkTxData struct {
	RssiPercent uint8
	Unknown1    uint8
	Unknown2    uint8
	PowerIndex  uint8
	PacketRate  uint8 //fps/10 (50hz = 0x05 or 5)
}

func UnmarshalLinkTx(data []byte) (LinkTxData, error) {
	d := LinkTxData{}
	if len(data) != LinkTxFrameLength {
		return d, ErrFrameLength
	}
	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}
	//TODO check correct type?

	d.RssiPercent = data[1]
	d.Unknown1 = data[2]
	d.Unknown2 = data[3]
	d.PowerIndex = data[4]
	d.PacketRate = data[5]

	return d, nil
}

func (d *LinkTxData) String() string {
	rate := int(d.PacketRate) * 10
	return fmt.Sprintf("RssiPercent: %d%% Unknown1: %d Unknown2: %d PacketRate: %dhz",
		d.RssiPercent,
		d.Unknown1,
		d.Unknown2,
		rate,
	)
}
