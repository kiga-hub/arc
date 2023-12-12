package micro

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Health 结构体
type Health struct {
	IsHealth bool `json:"health"`
}

// GetHealth 获取Health
//
//goland:noinspection GoUnusedExportedFunction
func GetHealth(address string) (*Health, error) {
	client := resty.New()
	health := &Health{}
	resp, err := client.SetTimeout(time.Second*3).R().
		SetHeader(http.CanonicalHeaderKey("User-Agent"), "kiga").
		SetResult(health).
		Get(address + urlHealth)
	if err != nil {
		return nil, err
	}
	code := resp.StatusCode()
	if code != http.StatusOK {
		return nil, fmt.Errorf("return code %d", code)
	}
	return health, nil
}

// GetStatus 获取Status
//
//goland:noinspection GoUnusedExportedFunction
func GetStatus(address string) (*Status, error) {
	client := resty.New()
	status := &Status{}
	resp, err := client.SetTimeout(time.Second*3).R().
		SetHeader(http.CanonicalHeaderKey("User-Agent"), "kiga").
		SetResult(status).
		Get(address + urlStatus)
	if err != nil {
		return nil, err
	}
	code := resp.StatusCode()
	if code != http.StatusOK {
		return nil, fmt.Errorf("return code %d", code)
	}
	return status, nil
}
