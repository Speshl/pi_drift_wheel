package crsf

const (
	BaudRate    = 420000
	NumChannels = 16
	MinValue    = 172 //987us - actual CRSF min is 0 with E.Limits on
	Value1k     = 191
	MidValue    = 992
	Value2k     = 1792
	MaxValue    = 1811 //2012us - actual CRSF max is 1984 with E.Limits on
	SpanValue   = MaxValue - MinValue

	PacketSize = 64 // max declared len is 62+DEST+LEN on top of that = 64
	PayloadLen = PacketSize - 4

	SyncByte = 0xC8

	AddressLength     = 1                                                    // length of ADDRESS field
	FrameLengthLength = 1                                                    // length of FRAMELENGTH field
	TypeLength        = 1                                                    // length of TYPE field
	CRCLength         = 1                                                    // length of CRC field
	CRCAndTypeLength  = TypeLength + CRCLength                               // length of TYPE and CRC fields combined
	ExtTypeCRCLength  = 2 + CRCAndTypeLength                                 // length of Extended Dest/Origin, TYPE and CRC fields combined
	NonPayloadLength  = AddressLength + FrameLengthLength + CRCAndTypeLength // combined length of all fields except payload

	GPSFrameSize      = 15
	BatterySensorSize = 8
	LinkStatsSize     = 10
	ChannelsSize      = 22
	AttitudeSize      = 6

	GPSType           = 0x02
	BatterySensorType = 0x08
	LinkStatsType     = 0x14
	OpenTxSyncType    = 0x10
	RadioIdType       = 0x3A
	ChannelsType      = 0x16
	AttitudeType      = 0x1E
	FlightModeType    = 0x21

	// Extended Header Frames, range: 0x28 to 0x96
	DevicePingType     = 0x28
	DeviceInfoType     = 0x29
	ParameterEntryType = 0x2B
	ParameterReadType  = 0x2C
	ParameterWriteType = 0x2D
	CommandType        = 0x32

	// MSP commands
	MspReqType   = 0x7A
	MspRespType  = 0x7B
	MspWriteType = 0x7C

	BroadcastAddress        = 0x00
	UsbAddress              = 0x10
	TbsCoreAddress          = 0x80
	Reserved1Address        = 0x8A
	CurrentSensorAddress    = 0xC0
	GpsAddress              = 0xC2
	BlackBoxAddress         = 0xC4
	FlightControllerAddress = 0xC8
	Reserved2Address        = 0xCA
	RaceTagAddress          = 0xCC
	RadioTransmitterAddress = 0xEA
	ReceiverAddress         = 0xEC
	TransmitterAddress      = 0xEE
)
