package mqtt

import (
	"encoding/json"
)

const (
	// GlobalModelQuery - 下发传感器查询消息
	GlobalModelQuery = "QUERY"
)

// QueryPayload - 查询
type QueryPayload struct {
	GlobalPayload
	Response chan *VersionPayload `json:"-"`
}

// NewQueryPayload -
func NewQueryPayload() *QueryPayload {
	q := &QueryPayload{}
	q.SetVersion("1.1")
	q.SetMode(GlobalModelQuery)
	q.SetClient(Broadcast)
	q.SetSensor(Broadcast)
	q.SetGroup("admin")
	return q
}

// Pack - 编码
func (p *QueryPayload) Pack() ([]byte, error) {
	return json.Marshal(p)
}
