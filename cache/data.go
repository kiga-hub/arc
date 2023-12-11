package cache

import (
	"math"
	"time"
)

const (
	toleranceCoefficient = float64(.05) // 两帧数据包之间时间差容忍系数
)

// IDataPoint -
type IDataPoint interface {
	GetID() uint64
	GetTime() int64
	GetSize() uint64
	Append(data IDataPoint)
	IsAppendable(data IDataPoint) bool
}

// DataPoint is a structure for holding data of a time point
type DataPoint struct {
	ID         uint64
	Time       time.Time
	SampleRate uint16
	Channel    int
	Data       []byte
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

// IsAppendable -
func (d *DataPoint) IsAppendable(a IDataPoint) bool {
	ap := a.(*DataPoint)
	if ap == nil ||
		ap.SampleRate != d.SampleRate ||
		ap.Channel != d.Channel ||
		ap.GetTime()-d.GetTime()-d.getDuration() > getDataSizeGap(ap) {
		return false
	}

	return true
}

// getPoints 获取采样点数
func (d *DataPoint) getPoints() int {
	if d.Channel == 0 {
		return -math.MaxInt
	}
	return len(d.Data) / 2 / d.Channel
}

// getDuration 单位 us
func (d *DataPoint) getDuration() int64 {
	if d.SampleRate == 0 {
		return -math.MaxInt64
	}
	return int64(float64(d.getPoints()) / float64(d.SampleRate) * 1e6)
}

// getDataSizeGap 采样点数计算出两帧之间的时间差,因为时间精度的问题，计算结果再乘以容忍系数
func getDataSizeGap(ap *DataPoint) int64 {
	gap := int64(float64(ap.getDuration()) * toleranceCoefficient)
	return gap
}
