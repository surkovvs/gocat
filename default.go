package shutdowner

var sdDefault *shutdown

func GetDefault() *shutdown {
	if sdDefault == nil {
		sdDefault = NewShutdown()
	}
	return sdDefault
}

func SetDefault(sd *shutdown) {
	sdDefault = sd
}
