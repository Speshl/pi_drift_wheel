package crsf

import (
	"fmt"

	"github.com/Speshl/pi_drift_wheel/crsf/frames"
)

type CRSFData struct {
	Channels frames.ChannelsData
	CRSFTelemetry
	//OpenTxSync frames.OpenTxSyncData
	// DevicePing      frames.DevicePingData
	// DeviceInfo      frames.DeviceInfoData
	// RequestSettings frames.RequestSettingsData
	// ParameterEntry  frames.ParameterEntryData
	// ParameterRead   frames.ParameterReadData
	// ParameterWrite  frames.ParameterWriteData
	// Command         frames.CommandData
	// RadioId         frames.RadioIdData
	// MspRequest      frames.MspRequestData
	// MspResponse     frames.MspResponseData
	// MspWrite        frames.MspWriteData
	// DisplayCommand  frames.DisplayCommandData
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
	return fmt.Sprintf("GPS: {%v}\nVario: {%v}\nBattery: {%v}\nBarometer: {%v}\nLinkStats: {%v}\nLinkRx: {%v}\nLinkTx: {%v}\nAttitude: {%v}\nFlightMode: {%v}",
		d.Gps,
		d.Vario,
		d.BatterySensor,
		d.Barometer,
		d.LinkStats,
		d.LinkRx,
		d.LinkTx,
		d.Attitude,
		d.FlightMode,
	)
}
