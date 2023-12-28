package frames

type RequestSettingsData struct {
}

func UnmarshalRequestSettings(data []byte) (RequestSettingsData, error) {
	return RequestSettingsData{}, nil
}
