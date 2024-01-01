package alertmanager

import (
	"time"

	"github.com/kiga-hub/arc/logging"
	"github.com/labstack/echo/v4"
)

// Alert -
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// Notification -
type Notification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

// WebhookHandler is an interface that wraps the initWebhookHandler method
type WebhookHandler interface {
	RegisterWebhookHandler(c echo.Context) error
}

// AlertManager -
type AlertManager struct {
	logger logging.ILogger
	config *Config
	handle WebhookHandler
}

// New -
func New(config *Config, logger logging.ILogger) *AlertManager {
	return &AlertManager{
		config: config,
		logger: logger,
	}
}

// InitWebhookHandler -
func (a *AlertManager) InitWebhookHandler(h WebhookHandler) {
	a.handle = h
}

// Start -
func (a *AlertManager) Start() {
}

// Close -
func (a *AlertManager) Close() {
}
