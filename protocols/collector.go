package protocols

// TemperatureArray is an array of TemperatureItem
type TemperatureArray struct {
	CollectorID   []byte
	CollectorType string
	Items         []TemperatureItem
}

// TemperatureItem is the temperature at a timepoint
type TemperatureItem struct {
	Ts   int64
	Data float32
}

// VibrateArray -
type VibrateArray struct {
	CollectorID   []byte
	CollectorType string
	Items         []VibrateValues
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

// VibrateValues -
type VibrateValues struct {
	Ts int64
	X  int16
	Y  int16
	Z  int16
}
