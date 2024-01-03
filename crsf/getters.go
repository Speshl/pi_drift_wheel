package crsf

import "github.com/Speshl/pi_drift_wheel/crsf/frames"

func (c *CRSF) GetData() CRSFData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data
}

func (c *CRSF) GetGps() frames.GpsData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.Gps
}

func (c *CRSF) GetVario() frames.VarioData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.Vario
}

func (c *CRSF) GetBatterySensor() frames.BatterySensorData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.BatterySensor
}

func (c *CRSF) GetBarometer() frames.BarometerData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.Barometer
}

func (c *CRSF) GetLinkStats() frames.LinkStatsData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.LinkStats
}

func (c *CRSF) GetChannels() frames.ChannelsData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.Channels
}

func (c *CRSF) GetLinkRx() frames.LinkRxData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.LinkRx
}

func (c *CRSF) GetLinkTx() frames.LinkTxData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.LinkTx
}

func (c *CRSF) GetAttitude() frames.AttitudeData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.Attitude
}

func (c *CRSF) GetFlightMode() frames.FlightModeData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	return c.data.FlightMode
}
