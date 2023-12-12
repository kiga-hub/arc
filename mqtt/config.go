package mqtt

import (
	"encoding/json"

	"github.com/kiga-hub/arc/protocols"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// GlobalModelConfig - 传感器配置消息
	GlobalModelConfig = "CONFIG"
	// GlobalModelAck - 上报传感器配置应答消息
	GlobalModelAck = "ACK"
)

// ConfigPayload - config 消息管理
type ConfigPayload struct {
	GlobalPayload
	ConfigItemSampleRate int `json:"CONFIG_ITEM_SAMPLERATE"`
}

// Validate - 校验
func (p *ConfigPayload) Validate() error {
	return validation.ValidateStruct(p,
		// 检查音频采样率
		validation.Field(&p.ConfigItemSampleRate, validation.Required,
			validation.In(
				int(protocols.ItemSampleRate),
			),
		),
	)
}

// Pack - 编码消息
func (p *ConfigPayload) Pack() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

/*
 ****************************************************
 *				config api
 ****************************************************
 */

// NewConfigPayload - 创建config消息
func NewConfigPayload() *ConfigPayload {
	cfg := &ConfigPayload{
		ConfigItemSampleRate: int(protocols.ItemSampleRate),
	}
	cfg.SetVersion("1.1")
	cfg.SetMode(GlobalModelConfig)
	cfg.SetClient(Broadcast)
	cfg.SetSensor(Broadcast)
	cfg.SetGroup("admin")
	return cfg
}
