package protocols

// ArcItemArray is an array of arc item
type ArcItemArray struct {
	CollectorID   []byte
	CollectorType string
	Items         []ArcItem
}

// ArcItem is the ArcItem at a time point
type ArcItem struct {
	Ts   int64
	Data []byte
}

// ArcItemRaw include raw data of ArcItem value
type ArcItemRaw struct {
	CollectorID   []byte
	CollectorType string
	Ts            int64
	Data          []byte
}

// CountItem contains count of messages from collector
type CountItem struct {
	CollectorID   []byte
	CollectorType string
	Ts            int64
	Count         int
	Size          int
}
