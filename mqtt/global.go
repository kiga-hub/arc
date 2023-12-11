package mqtt

const (
	// Broadcast -
	Broadcast = "*"

	// TerminalSensorID -
	TerminalSensorID = "FFFFFFFFFFFF"
)

// IPayload -
type IPayload interface {
	Pack() ([]byte, error)
	GetClientID() string
}

// GlobalPayload - 公共消息
type GlobalPayload struct {
	GlobalVersion  string `json:"GLOBAL_VERSION"`  // 版本
	GlobalMode     string `json:"GLOBAL_MODE"`     // 模块
	GlobalClientID string `json:"GLOBAL_CLIENTID"` // 终端
	GlobalSensorID string `json:"GLOBAL_SENSORID"` // 传感器
	GlobalGroupID  string `json:"GLOBAL_GROUPID"`  // 分组
}

// SetVersion - 设置版本
func (p *GlobalPayload) SetVersion(id string) {
	p.GlobalVersion = id
}

// SetMode - 设置模块
func (p *GlobalPayload) SetMode(mode string) {
	p.GlobalMode = mode
}

// SetClient - 设置终端
func (p *GlobalPayload) SetClient(id string) {
	p.GlobalClientID = id
}

// SetSensor - 设置传感器
func (p *GlobalPayload) SetSensor(id string) {
	p.GlobalSensorID = id
}

// SetGroup - 分组(保留)
func (p *GlobalPayload) SetGroup(id string) {
	p.GlobalGroupID = id
}

// GetClientID -
func (p *GlobalPayload) GetClientID() string {
	return p.GlobalClientID
}
