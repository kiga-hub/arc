package cache

import (
	"time"
)

// IDataPoint -
type IDataPoint interface {
	GetID() uint64
	GetTime() int64
	GetSize() uint64
	Append(data IDataPoint)
}

// DataPoint is a structure for holding data of a time point
type DataPoint struct {
	ID   uint64
	Time time.Time
	Data []byte
}

// GetID -
func (d *DataPoint) GetID() uint64 {
	return d.ID
}

// GetTime -
func (d *DataPoint) GetTime() int64 {
	return d.Time.UnixMicro()
}

// GetSize -
func (d *DataPoint) GetSize() uint64 {
	return uint64(len(d.Data))
}

// Append -
func (d *DataPoint) Append(data IDataPoint) {
	dp := data.(*DataPoint)
	if dp == nil {
		return
	}

	d.Data = append(d.Data, dp.Data...)
}

// getPoints 获取采样点数
func (d *DataPoint) getPoints() int {
	return len(d.Data)
}
