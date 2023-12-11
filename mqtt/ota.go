package mqtt

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// GlobalModelOta - 传感器固件升级消息
	GlobalModelOta = "OTA"
)

// OtaPayload - 升级消息
type OtaPayload struct {
	GlobalPayload
	OtaURL string `json:"OTA_URL"`
	Md5    string `json:"MD5"`
	Size   int64  `json:"SIZE"`
}

// Validate - 校验
func (p *OtaPayload) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.OtaURL, validation.Required),
	)
}

// Pack - 编码
func (p *OtaPayload) Pack() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// SetURL -
func (p *OtaPayload) SetURL(url string) {
	p.OtaURL = url
}

// NewOtaPayload - 创建消息
func NewOtaPayload() *OtaPayload {
	ota := &OtaPayload{}
	ota.SetVersion("1.1")
	ota.SetMode(GlobalModelOta)
	ota.SetClient(Broadcast)
	ota.SetSensor(Broadcast)
	ota.SetGroup("admin")
	return ota
}

const (
	// OtaNoticeAckStage -
	OtaNoticeAckStage = "ACK"
	// OtaNoticeCheckStage -
	OtaNoticeCheckStage = "CHECK"
	// OtaNoticeOkStatus -
	OtaNoticeOkStatus = "OK"
	// OtaNoticeFailedStatus -
	OtaNoticeFailedStatus = "FAILED"
)

// OtaNoticePayload -
type OtaNoticePayload struct {
	GlobalPayload
	Stage  string `json:"STAGE"`
	Status string `json:"STATUS"`
	Reason int64  `json:"REASON"`
}
