package protocols

// TemperatureArray is an array of TemperatureItem
type TemperatureArray struct {
	CollectorID   []byte
	CollectorType string
	Items         []TemperatureItem
}

// TemperatureItem is the temperature at a time point
type TemperatureItem struct {
	Ts   int64
	Data float32
}

// TemperatureRaw include raw data of temperature value
type TemperatureRaw struct {
	CollectorID   []byte
	CollectorType string
	Ts            int64
	Data          float32
}

// CountItem contains count of messages from collector
type CountItem struct {
	CollectorID   []byte
	CollectorType string
	Ts            int64
	Count         int
	Size          int
}
