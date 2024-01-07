// https://github.com/crsf-wg/crsf/wiki/CRSF_FRAMETYPE_GPS
package frames

import (
	"encoding/binary"
	"fmt"
)

const (
	GpsFrameLength = 15 + 2 //Payload + Type + CRC
)

type GpsData struct {
	Lat            int32 //latitude in degress * 10000000, big-endian
	Long           int32
	Speed          int16  //km/h * 10, big-endian
	Course         int16  //gps heading in degress * 100, big-endian
	Altitude       uint16 //gps altitude in meters + 1000 m, big-endian
	SatelliteCount uint8
}

func UnmarshalGps(data []byte) (GpsData, error) {
	d := GpsData{}
	if len(data) != GpsFrameLength {
		return d, ErrFrameLength
	}

	if !ValidateFrame(data) {
		return d, ErrInvalidCRC8
	}

	//TODO check correct type?

	d.Lat = int32(binary.BigEndian.Uint32(data[1:5]))
	d.Long = int32(binary.BigEndian.Uint32(data[5:9]))
	d.Speed = int16(binary.BigEndian.Uint16(data[9:11]))
	d.Course = int16(binary.BigEndian.Uint16(data[11:13]))
	d.Altitude = binary.BigEndian.Uint16(data[13:15])
	d.SatelliteCount = data[15]

	return d, nil
}

func (d *GpsData) String() string {
	lat := float32(d.Lat) / 10000000
	long := float32(d.Long) / 10000000
	speed := float32(d.Speed) / 10
	course := float32(d.Course) / 100
	altitude := float32(d.Altitude) - 1000
	satCount := d.SatelliteCount

	return fmt.Sprintf("Lat: %.7f Long: %.7f Speed: %.1fkph Course: %.2f Altitude: %.0fm SatCount: %d", lat, long, speed, course, altitude, satCount)
}
