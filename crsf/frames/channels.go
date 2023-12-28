// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_RC_CHANNELS_PACKED
package frames

import (
	"fmt"
)

const (
	ChannelsFrameLength        = 22 + 2 //Payload + Type + CRC
	MaxChannels                = 16
	ChannelsMask        uint16 = 0x07ff // The maximum 11-bit channel value

	//(0-1984)
	ChannelsMin = 172  //988us
	ChannelsMid = 992  //1500us
	ChannelsMax = 1811 //2012us
)

type ChannelsData struct {
	Channels []uint16
}

func UnmarshalChannels(data []byte) (ChannelsData, error) {
	d := ChannelsData{
		Channels: make([]uint16, MaxChannels),
	}

	if len(data) != ChannelsFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Channels[0] = ((uint16(data[1]) | uint16(data[2])<<8) & ChannelsMask)
	d.Channels[1] = ((uint16(data[2])>>3 | uint16(data[3])<<5) & ChannelsMask)
	d.Channels[2] = ((uint16(data[3])>>6 | uint16(data[4])<<2 | uint16(data[5])<<10) & ChannelsMask)
	d.Channels[3] = ((uint16(data[5])>>1 | uint16(data[6])<<7) & ChannelsMask)
	d.Channels[4] = ((uint16(data[6])>>4 | uint16(data[7])<<4) & ChannelsMask)
	d.Channels[5] = ((uint16(data[7])>>7 | uint16(data[8])<<1 | uint16(data[9])<<9) & ChannelsMask)
	d.Channels[6] = ((uint16(data[9])>>2 | uint16(data[10])<<6) & ChannelsMask)
	d.Channels[7] = ((uint16(data[10])>>5 | uint16(data[11])<<3) & ChannelsMask)
	d.Channels[8] = ((uint16(data[12]) | uint16(data[13])<<8) & ChannelsMask)
	d.Channels[9] = ((uint16(data[13])>>3 | uint16(data[14])<<5) & ChannelsMask)
	d.Channels[10] = ((uint16(data[14])>>6 | uint16(data[15])<<2 | uint16(data[16])<<10) & ChannelsMask)
	d.Channels[11] = ((uint16(data[16])>>1 | uint16(data[17])<<7) & ChannelsMask)
	d.Channels[12] = ((uint16(data[17])>>4 | uint16(data[18])<<4) & ChannelsMask)
	d.Channels[13] = ((uint16(data[18])>>7 | uint16(data[19])<<1 | uint16(data[20])<<9) & ChannelsMask)
	d.Channels[14] = ((uint16(data[20])>>2 | uint16(data[21])<<6) & ChannelsMask)
	d.Channels[15] = ((uint16(data[21])>>5 | uint16(data[22])<<3) & ChannelsMask)

	//TODO CRC byte?
	return d, nil
}

func (d *ChannelsData) String() string {
	builtString := ""
	for i := range d.Channels {
		if i != 0 {
			builtString = fmt.Sprintf("%s Channel%d: %d", builtString, i+1, d.Channels[i])
		} else {
			builtString = fmt.Sprintf("Channel%d: %d", i+1, d.Channels[i])
		}
	}

	return builtString
}
