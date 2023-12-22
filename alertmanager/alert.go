package alertmanager

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kiga-hub/arc/logging"
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
	RegisterWebhookHandler(w http.ResponseWriter, r *http.Request)
	WsConnect(w http.ResponseWriter, r *http.Request)
}

// AlertManager -
type AlertManager struct {
	logger logging.ILogger
	config *Config
	server *http.Server
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
	address := fmt.Sprintf("%s:%d", a.config.ServerAddress, a.config.ServerPort)

	mux := http.NewServeMux()
	mux.HandleFunc(a.config.URL, a.handle.RegisterWebhookHandler)
	mux.HandleFunc("/ws", a.handle.WsConnect)

	a.server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	fmt.Println("http.Start :", a.server)

	go func() {
		if err := a.server.ListenAndServe(); err != nil {
			fmt.Println("http.ListenAndServe失败:", err)
			return
		}
	}()

}

// Close -
func (a *AlertManager) Close() {
	// close http
	if a.server != nil {
		if err := a.server.Shutdown(context.Background()); err != nil {
			a.logger.Error("Failed to shutdown server:", err)
		}
	}
}
