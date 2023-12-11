package mqtt

import "encoding/json"

const (
	// GlobalModelVersion - 传感器上报版本消息
	GlobalModelVersion = "VERSION"
)

// VersionPayload - 配置
type VersionPayload struct {
	GlobalPayload
	VersionHw      string `json:"VERSION_HW"`
	VersionSw      string `json:"VERSION_SW"`
	CollectorModel string `json:"COLLECTOR_MODEL"`
}

// NewVersionPayload -
func NewVersionPayload() *VersionPayload {
	version := &VersionPayload{}
	version.SetVersion("1.1")
	version.SetMode(GlobalModelVersion)
	version.SetClient(Broadcast)
	version.SetSensor(Broadcast)
	version.SetGroup("admin")
	return version
}

// Pack -
func (p *VersionPayload) Pack() ([]byte, error) {
	return json.Marshal(p)
}
