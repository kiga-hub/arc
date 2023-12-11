package mqtt

import (
	"encoding/json"

	"github.com/kiga-hub/common/protocols"

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
	ConfigItemASamplerate int `json:"CONFIG_ITEM_A_SAMPLERATE"`
	ConfigItemABits       int `json:"CONFIG_ITEM_A_BITS"`
	ConfigItemAChannel    int `json:"CONFIG_ITEM_A_CHANNEL"`
	ConfigItemVSamplerate int `json:"CONFIG_ITEM_V_SAMPLERATE"`
	ConfigItemVRange      int `json:"CONFIG_ITEM_V_RANGE"`
	ConfigItemVResolution int `json:"CONFIG_ITEM_V_RESOLUTION"`
	ConfigItemVOffsetX    int `json:"CONFIG_ITEM_V_OFFSET_X"`
	ConfigItemVOffsetY    int `json:"CONFIG_ITEM_V_OFFSET_Y"`
	ConfigItemVOffsetZ    int `json:"CONFIG_ITEM_V_OFFSET_Z"`
}

// Validate - 校验
func (p *ConfigPayload) Validate() error {
	return validation.ValidateStruct(p,
		// 检查音频采样率
		validation.Field(&p.ConfigItemASamplerate, validation.Required,
			validation.In(
				int(protocols.AudioSampleRate8k),
				int(protocols.AudioSampleRate16k),
				int(protocols.AudioSampleRate32k),
				int(protocols.AudioSampleRate48k),
			),
		),
		// 检查音频位深度
		validation.Field(&p.ConfigItemABits, validation.Required,
			validation.In(8, 16, 32),
		),
		// 检查音频通道数
		validation.Field(&p.ConfigItemAChannel, validation.Required,
			validation.Min(1),
			validation.Max(8),
		),
		// 检查震动采样率
		validation.Field(&p.ConfigItemVSamplerate,
			validation.Min(0),
			validation.Max(15),
		),
		// 检查震动量程范围
		validation.Field(&p.ConfigItemVRange, validation.Required,
			validation.In(2, 4, 8, 16),
		),
		// 检查震动分辨率
		validation.Field(&p.ConfigItemVResolution, validation.Required,
			validation.In(10, 11, 12, 13),
		),
	)
}

// Pack - 编码消息
func (p *ConfigPayload) Pack() ([]byte, error) {
	if p.ConfigItemVResolution == int(protocols.VibrateResolution11) {
		p.ConfigItemVRange = int(protocols.VibrateDynamicRange4)
	} else if p.ConfigItemVResolution == int(protocols.VibrateResolution12) {
		p.ConfigItemVRange = int(protocols.VibrateDynamicRange8)
	} else if p.ConfigItemVResolution == int(protocols.VibrateResolution13) {
		p.ConfigItemVRange = int(protocols.VibrateDynamicRange16)
	}
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
		ConfigItemASamplerate: int(protocols.AudioSampleRate48k),
		ConfigItemABits:       16,
		ConfigItemAChannel:    1,
		ConfigItemVSamplerate: protocols.VibrateSampleRate3200,
		ConfigItemVRange:      int(protocols.VibrateDynamicRange2),
		ConfigItemVResolution: int(protocols.VibrateResolution10),
	}
	cfg.SetVersion("1.1")
	cfg.SetMode(GlobalModelConfig)
	cfg.SetClient(Broadcast)
	cfg.SetSensor(Broadcast)
	cfg.SetGroup("admin")
	return cfg
}
