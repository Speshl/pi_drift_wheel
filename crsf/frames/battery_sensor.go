// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_VARIO
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	BatterySensorFrameLength = 8 + 2 //Payload + Type + CRC
)

type BatterySensorData struct {
	Voltage   int16 // dv Big-Endian
	Current   int16 // da Big Endian
	Used      int32 //int24 mAh Little Endian
	Remaining int8  //percent (0-100)
}

func UnmarshalBatterySensor(data []byte) (BatterySensorData, error) {
	d := BatterySensorData{}
	if len(data) != BatterySensorFrameLength {
		return d, ErrFrameLength
	}

	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}
	//TODO check correct type?
	d.Voltage = int16(binary.BigEndian.Uint16(data[1:3]))
	d.Current = int16(binary.BigEndian.Uint16(data[3:5]))

	paddedInt24 := []byte{0} //add zero byte to front of 3 data bytes to pad out to 4 bytes for int32, little endian so 0 should go first
	paddedInt24 = append(paddedInt24, data[5:8]...)
	d.Used = int32(binary.LittleEndian.Uint32(paddedInt24))

	d.Remaining = int8(data[8])

	//TODO CRC byte?
	return d, nil
}

func (d *BatterySensorData) String() string {
	voltage := int(d.Voltage) * 10
	return fmt.Sprintf("Voltage: %dV Current: %dda Used: %dmAh Remaining: %d%%", voltage, d.Current, d.Used, d.Remaining)
}
