package mqtt

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ActionPowerOn - 启动发送数据
	ActionPowerOn = "ON"
	// ActionPowerOff - 关闭发送数据
	ActionPowerOff = "OFF"
	// ActionPowerReset - 传感器重启
	ActionPowerReset = "RESET"
)

const (
	// GlobalModelAction - 传感器动作消息
	GlobalModelAction = "ACTION"
)

// ActionPayload - action 消息管理
type ActionPayload struct {
	GlobalPayload
	ActionPower string `json:"ACTION_POWER"`
}

// Validate - 校验
func (p *ActionPayload) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.ActionPower, validation.Required,
			validation.In(
				ActionPowerOn, ActionPowerOff, ActionPowerReset,
			),
		),
	)
}

// Pack - 消息编码
func (p *ActionPayload) Pack() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// SetPowerOn -
func (p *ActionPayload) SetPowerOn() {
	p.ActionPower = ActionPowerOn
}

// SetPowerOff -
func (p *ActionPayload) SetPowerOff() {
	p.ActionPower = ActionPowerOff
}

// SetPowerReset -
func (p *ActionPayload) SetPowerReset() {
	p.ActionPower = ActionPowerReset
}

/*
 ****************************************************
 *				action api
 ****************************************************
 */

// NewActionPayload - 创建消息结构
func NewActionPayload() *ActionPayload {
	action := &ActionPayload{}
	action.SetVersion("1.1")
	action.SetMode(GlobalModelAction)
	action.SetClient(Broadcast)
	action.SetSensor(Broadcast)
	action.SetGroup("admin")
	action.ActionPower = ActionPowerOff
	return action
}
