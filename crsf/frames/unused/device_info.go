// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_DEVICE_INFO
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	DeviceInfoFrameLength = 2 + 2 + 2 //Payload + Type + CRC + dest + src
)

type DeviceInfoData struct {
	Destination uint8
	Source      uint8

	Name         string
	Serial       uint32
	HWVersion    uint32
	SWVersion    uint32
	ParamCount   uint8
	ProtoVersion uint8
}

func UnmarshalDeviceInfo(data []byte) (DeviceInfoData, error) {
	d := DeviceInfoData{}
	if len(data) != DeviceInfoFrameLength {
		return d, fmt.Errorf("incorrect frame length")
	}
	//TODO check correct type?

	d.Destination = data[1]
	d.Source = data[2]

	i := 3
	for i = 3; i < len(data)-1; i++ {
		if i == 0x00 { //null terminator for string
			break
		}
		d.Name += string(data[i])
	}

	nextByte := i + 1

	d.Serial = binary.LittleEndian.Uint32(data[nextByte : nextByte+4])
	d.HWVersion = binary.LittleEndian.Uint32(data[nextByte+4 : nextByte+8])
	d.SWVersion = binary.LittleEndian.Uint32(data[nextByte+8 : nextByte+12])
	d.ParamCount = data[nextByte+12]
	d.ProtoVersion = data[nextByte+13]

	//TODO CRC byte?
	return d, nil
}
