package frames

type OpenTxSyncData struct {
}

func UnmarshalOpenTxSync(data []byte) (OpenTxSyncData, error) {
	return OpenTxSyncData{}, nil
}
