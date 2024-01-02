package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/kiga-hub/arc/alertmanager"
	"github.com/kiga-hub/arc/micro"
	"github.com/kiga-hub/arc/micro/component"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

// MyWebhookHandler -
type MyWebhookHandler struct {
}

// RegisterWebhookHandler -
func (m *MyWebhookHandler) RegisterWebhookHandler(c echo.Context) error {

	req := c.Request()
	if req.Method != http.MethodPost {
		return c.String(http.StatusMethodNotAllowed, "Invalid request method")
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var notice alertmanager.Notification
	err = json.Unmarshal(body, &notice)
	if err != nil {
		return err
	}

	// Handle the webhook
	err = HandleWebhook(notice)
	if err != nil {
		return err
	}

	for _, v := range notice.Alerts {
		fmt.Printf("v: %+v\n", v)
	}
	return nil
}

var (
	handleFunc = &MyWebhookHandler{}
)

// HandleWebhook - TODO
func HandleWebhook(notice alertmanager.Notification) error {
	version := notice.Version
	groupKey := notice.GroupKey
	status := notice.Status
	receiver := notice.Receiver
	groupLabels := notice.GroupLabels
	commonLabels := notice.CommonLabels
	commonAnnotations := notice.CommonAnnotations
	externalURL := notice.ExternalURL
	alerts := notice.Alerts

	fmt.Println("++++++++++++++++++++++++++++++++++++++++")
	formattedOutput := fmt.Sprintf("Notification received:\nVersion: %s\nGroupKey: %s\nStatus: %s\nReceiver: %s\nGroupLabels: %v\nCommonLabels: %v\nCommonAnnotations: %v\nExternalURL: %s\nAlerts: %v",
		version, groupKey, status, receiver, groupLabels, commonLabels, commonAnnotations, externalURL, alerts)
	fmt.Println(formattedOutput)

	return nil
}

func main() {

	go sendToPrometheus()

	server, err := micro.NewServer(
		"demo",
		"v1",
		[]micro.IComponent{
			&component.LoggingComponent{},
			// &tracing.Component{},
			&component.GossipKVCacheComponent{
				ClusterName:   "platform-global",
				Port:          6666,
				InMachineMode: false,
			},
			// &kafka.Component{},
			&alertmanager.Component{},
			// &taos.Component{},
		},
	)

	if err != nil {
		panic(err)
	}
	err = server.Init()
	if err != nil {
		panic(err)
	}

	// 获取aleertmanager组件
	am := server.GetElement(&alertmanager.ElementKey).(*alertmanager.AlertManager)
	am.InitWebhookHandler(handleFunc)

	err = server.Run()
	if err != nil {
		fmt.Println(err)
	}

}

var (
	gauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "app",
		Subsystem: "process",
		Name:      "frames_ch_blocked_num",
	})
)

func init() {
	prometheus.MustRegister(
		gauge,
	)
}
func sendToPrometheus() {

	for {
		time.Sleep(time.Second)
		gauge.Set(float64(rand.Intn(120)))
	}

}
