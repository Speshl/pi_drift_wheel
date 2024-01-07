// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_LINK_STATISTICS
package frames

import (
	"fmt"
)

const (
	LinkStatsFrameLength = 10 + 2 //Payload + Type + CRC
)

type LinkStatsData struct {
	UplinkRssiAnt1     uint8 //dBm * -1
	UplinkRssiAnt2     uint8 //dBm * -1
	UplinkQuality      uint8 // (0-100)%
	UplinkSnr          int8  //db
	DiversifyActiveAnt uint8 //( enum ant. 1 = 0, ant. 2 = 1 )
	RfMode             uint8 //(500hz, 250hz etc... varies)
	Power              uint8 // ( enum 0mW = 0, 10mW, 25 mW, 100 mW, 500 mW, 1000 mW, 2000mW, 50mW )
	DownlinkRssi       uint8 //dBm * -1
	DownlinkQuality    uint8 // (0-100)%
	DownlinkSnr        uint8 //db
}

func UnmarshalLinkStats(data []byte) (LinkStatsData, error) {
	d := LinkStatsData{}
	if len(data) != LinkStatsFrameLength {
		return d, ErrFrameLength
	}

	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}
	//TODO check correct type?

	d.UplinkRssiAnt1 = data[1]
	d.UplinkRssiAnt2 = data[2]
	d.UplinkQuality = data[3]
	d.UplinkSnr = int8(data[4])
	d.DiversifyActiveAnt = data[5]
	d.RfMode = data[6]
	d.Power = data[7]
	d.DownlinkRssi = data[8]
	d.DownlinkQuality = data[9]
	d.DownlinkSnr = data[10]
	return d, nil
}

func (d *LinkStatsData) String() string {
	txRssiAnt1 := int8(d.UplinkRssiAnt1) * -1
	txRssiAnt2 := int8(d.UplinkRssiAnt2) * -1

	activeAnt := d.DiversifyActiveAnt + 1

	power := 0
	switch d.Power {
	case 0:
		power = 0
	case 1:
		power = 10
	case 2:
		power = 25
	case 3:
		power = 100
	case 4:
		power = 500
	case 5:
		power = 1000
	case 6:
		power = 2000
	case 7:
		power = 50
	}

	rxRssi := int8(d.DownlinkRssi) * -1

	return fmt.Sprintf("TxRssiAnt1: %ddBm TxRssiAnt2: %ddBm TxQuality: %d%% TxSNR: %ddb ActiveAnt: %d RFMode: %dhz Power: %dmw RxRSSI: %ddBm RxQuality: %d%% RxSNR: %ddb",
		txRssiAnt1,
		txRssiAnt2,
		d.UplinkQuality,
		d.UplinkSnr,
		activeAnt,
		d.RfMode,
		power,
		rxRssi,
		d.DownlinkQuality,
		d.DownlinkSnr,
	)
}
