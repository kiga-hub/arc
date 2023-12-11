package mqtt

const (
	// NoticeDisconnected - 传感器离线
	NoticeDisconnected = "sensor_disconnected"
	// NoticeConnected - 传感器上线
	NoticeConnected = "sensor_connected"
)

const (
	// GlobalModelNotice - 传感器上报通知消息
	GlobalModelNotice = "NOTICE"
)

// NoticePayload - 查询
type NoticePayload struct {
	GlobalPayload
	Action string `json:"ACTION"`
	Reason string `json:"REASON"`
}
