package crsf

import (
	"fmt"

	"github.com/Speshl/pi_drift_wheel/crsf/frames"
)

type CRSFData struct {
	Channels frames.ChannelsData
	CRSFTelemetry
}

type CRSFTelemetry struct {
	Gps           frames.GpsData
	Vario         frames.VarioData
	BatterySensor frames.BatterySensorData
	Barometer     frames.BarometerData
	LinkStats     frames.LinkStatsData
	LinkRx        frames.LinkRxData
	LinkTx        frames.LinkTxData
	Attitude      frames.AttitudeData
	FlightMode    frames.FlightModeData
}

func NewCRSFData() CRSFData {
	return CRSFData{}
}

func (d *CRSFData) String() string {
	return fmt.Sprintf("Channels: {%s}\nGPS: {%s}\nVario: {%s}\nBattery: {%s}\nBarometer: {%s}\nLinkStats: {%s}\nLinkRx: {%s}\nLinkTx: {%s}\nAttitude: {%s}\nFlightMode: {%s}",
		d.Channels.String(),
		d.Gps.String(),
		d.Vario.String(),
		d.BatterySensor.String(),
		d.Barometer.String(),
		d.LinkStats.String(),
		d.LinkRx.String(),
		d.LinkTx.String(),
		d.Attitude.String(),
		d.FlightMode.String(),
	)
}
