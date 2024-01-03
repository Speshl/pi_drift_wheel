package crsf

import (
	"log/slog"

	"github.com/Speshl/pi_drift_wheel/crsf/frames"
)

func (c *CRSF) SetData(data CRSFData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data = data
}

func (c *CRSF) updateGps(data []byte) error {
	dataStruct, err := frames.UnmarshalGps(data)
	if err != nil {
		return err
	}
	c.SetGps(dataStruct)
	return nil
}

func (c *CRSF) SetGps(data frames.GpsData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.Gps = data
}

func (c *CRSF) updateVario(data []byte) error {
	dataStruct, err := frames.UnmarshalVario(data)
	if err != nil {
		return err
	}
	c.SetVario(dataStruct)
	return nil
}

func (c *CRSF) SetVario(data frames.VarioData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.Vario = data
}

func (c *CRSF) updateBatterySensor(data []byte) error {
	dataStruct, err := frames.UnmarshalBatterySensor(data)
	if err != nil {
		return err
	}
	c.SetBatterySensor(dataStruct)
	return nil
}

func (c *CRSF) SetBatterySensor(data frames.BatterySensorData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.BatterySensor = data
}

func (c *CRSF) updateBarometer(data []byte) error {
	dataStruct, err := frames.UnmarshalBarometer(data)
	if err != nil {
		return err
	}
	c.SetBarometer(dataStruct)
	return nil
}

func (c *CRSF) SetBarometer(data frames.BarometerData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.Barometer = data
}

// func (c *CRSF) updateOpenTxSync(data []byte) error {
// 	dataStruct, err := frames.UnmarshalOpenTxSync(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetOpenTxSync(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetOpenTxSync(data frames.OpenTxSyncData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.OpenTxSync = data
// }

func (c *CRSF) updateLinkStats(data []byte) error {
	dataStruct, err := frames.UnmarshalLinkStats(data)
	if err != nil {
		return err
	}
	c.SetLinkStats(dataStruct)
	return nil
}

func (c *CRSF) SetLinkStats(data frames.LinkStatsData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.LinkStats = data
}

func (c *CRSF) updateChannels(data []byte) error {
	dataStruct, err := frames.UnmarshalChannels(data)
	if err != nil {
		return err
	}
	c.SetChannels(dataStruct)
	return nil
}

func (c *CRSF) SetChannels(data frames.ChannelsData) {
	slog.Info("setting channels", "data", data.String())
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.Channels = data
}

func (c *CRSF) updateLinkRx(data []byte) error {
	dataStruct, err := frames.UnmarshalLinkRx(data)
	if err != nil {
		return err
	}
	c.SetLinkRx(dataStruct)
	return nil
}

func (c *CRSF) SetLinkRx(data frames.LinkRxData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.LinkRx = data
}

func (c *CRSF) updateLinkTx(data []byte) error {
	dataStruct, err := frames.UnmarshalLinkTx(data)
	if err != nil {
		return err
	}
	c.SetLinkTx(dataStruct)
	return nil
}

func (c *CRSF) SetLinkTx(data frames.LinkTxData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.LinkTx = data
}

func (c *CRSF) updateAttitude(data []byte) error {
	dataStruct, err := frames.UnmarshalAttitude(data)
	if err != nil {
		return err
	}
	c.SetAttitude(dataStruct)
	return nil
}

func (c *CRSF) SetAttitude(data frames.AttitudeData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.Attitude = data
}

func (c *CRSF) updateFlightMode(data []byte) error {
	dataStruct, err := frames.UnmarshalFlightMode(data)
	if err != nil {
		return err
	}
	c.SetFlightMode(dataStruct)
	return nil
}

func (c *CRSF) SetFlightMode(data frames.FlightModeData) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.data.FlightMode = data
}

// func (c *CRSF) updateDevicePing(data []byte) error {
// 	dataStruct, err := frames.UnmarshalDevicePing(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetDevicePing(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetDevicePing(data frames.DevicePingData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.DevicePing = data
// }

// func (c *CRSF) updateDeviceInfo(data []byte) error {
// 	dataStruct, err := frames.UnmarshalDeviceInfo(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetDeviceInfo(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetDeviceInfo(data frames.DeviceInfoData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.DeviceInfo = data
// }

// func (c *CRSF) updateRequestSettings(data []byte) error {
// 	dataStruct, err := frames.UnmarshalRequestSettings(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetRequestSettings(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetRequestSettings(data frames.RequestSettingsData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.RequestSettings = data
// }

// func (c *CRSF) updateParameterEntry(data []byte) error {
// 	dataStruct, err := frames.UnmarshalParameterEntry(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetParameterEntry(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetParameterEntry(data frames.ParameterEntryData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.ParameterEntry = data
// }

// func (c *CRSF) updateParameterRead(data []byte) error {
// 	dataStruct, err := frames.UnmarshalParameterRead(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetParameterRead(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetParameterRead(data frames.ParameterReadData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.ParameterRead = data
// }

// func (c *CRSF) updateParameterWrite(data []byte) error {
// 	dataStruct, err := frames.UnmarshalParameterWrite(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetParameterWrite(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetParameterWrite(data frames.ParameterWriteData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.ParameterWrite = data
// }

// func (c *CRSF) updateCommand(data []byte) error {
// 	dataStruct, err := frames.UnmarshalCommand(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetCommand(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetCommand(data frames.CommandData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.Command = data
// }

// func (c *CRSF) updateRadioId(data []byte) error {
// 	dataStruct, err := frames.UnmarshalRadioId(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetRadioId(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetRadioId(data frames.RadioIdData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.RadioId = data
// }

// func (c *CRSF) updateMspRequest(data []byte) error {
// 	dataStruct, err := frames.UnmarshalMspRequest(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.MspRequest = dataStruct
// 	return nil
// }

// func (c *CRSF) SetMspRequest(data frames.LinkTxData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.LinkTx = data
// }

// func (c *CRSF) updateMspResponse(data []byte) error {
// 	dataStruct, err := frames.UnmarshalMspResponse(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetMspResponse(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetMspResponse(data frames.MspResponseData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.MspResponse = data
// }

// func (c *CRSF) updateMspWrite(data []byte) error {
// 	dataStruct, err := frames.UnmarshalMspWrite(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.MspWrite = dataStruct
// 	return nil
// }

// func (c *CRSF) SetMspWrite(data frames.MspWriteData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.MspWrite = data
// }

// func (c *CRSF) updateDisplayCommand(data []byte) error {
// 	dataStruct, err := frames.UnmarshalDisplayCommand(data)
// 	if err != nil {
// 		return err
// 	}
// 	c.SetDisplayCommand(dataStruct)
// 	return nil
// }

// func (c *CRSF) SetDisplayCommand(data frames.DisplayCommandData) {
// 	c.dataLock.Lock()
// 	defer c.dataLock.Unlock()
// 	c.data.DisplayCommand = data
// }
