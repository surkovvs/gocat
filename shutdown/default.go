package shutdown

var sdDefault *sdManager

func GetDefault() *sdManager {
	if sdDefault == nil {
		sdDefault = NewShutdown()
	}
	return sdDefault
}

func SetDefault(sd *sdManager) {
	sdDefault = sd
}
