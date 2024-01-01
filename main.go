package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kiga-hub/arc/alertmanager"
	"github.com/kiga-hub/arc/alertmanager/ws"
	"github.com/kiga-hub/arc/micro"
	"github.com/kiga-hub/arc/micro/component"
	"github.com/prometheus/client_golang/prometheus"
)

type wsMessageDataLabels struct {
	Job       string `json:"job"`
	Alertname string `json:"alertname"`
	Level     string `json:"level"`
	Instance  string `json:"instance"`
}
type wsMessageDataAnnotations struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Value       string `json:"value"`
}
type wsMessageData struct {
	Labels      wsMessageDataLabels      `json:"labels"`
	Annotations wsMessageDataAnnotations `json:"annotations"`
	T           int64                    `json:"t"`
}
type wsMessage struct {
	Type string        `json:"type"`
	Data wsMessageData `json:"data"`
}

// MyWebhookHandler -
type MyWebhookHandler struct {
	wsServer *ws.Server
}

// RegisterWebhookHandler -
func (m *MyWebhookHandler) RegisterWebhookHandler(c echo.Contex) error {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	// 	return
	// }

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "can't read body", http.StatusBadRequest)
	// 	return
	// }
	// get the data
	body, err := io.ReadAll(c.Request().Body())
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return err
	}

	var notice alertmanager.Notification
	err = json.Unmarshal(body, &notice)
	if err != nil {
		http.Error(w, "can't unmarshal JSON", http.StatusBadRequest)
		return err
	}

	// Handle the webhook
	err = HandleWebhook(notice)
	if err != nil {
		http.Error(w, "can't handle webhook", http.StatusInternalServerError)
		return
	}

	for _, v := range notice.Alerts {
		pushData := &wsMessage{
			Type: "alert",
			Data: wsMessageData{
				Labels: wsMessageDataLabels{
					Job:       v.Labels["job"],
					Alertname: v.Labels["alertname"],
					Level:     v.Labels["level"],
					Instance:  v.Labels["instance"],
				},
				Annotations: wsMessageDataAnnotations{
					Summary:     v.Annotations["summary"],
					Description: v.Annotations["description"],
					Value:       v.Annotations["value"],
				},
				T: v.StartsAt.Unix(),
			},
		}
		byteData, err := json.Marshal(pushData)
		if err != nil {
			http.Error(w, "can't unmarshal JSON", http.StatusBadRequest)
			return err
		}
		m.wsServer.Broadcast(byteData)

	}
	return nil
}

// WsConnect -
func (m *MyWebhookHandler) WsConnect(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var socketUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := socketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := ws.NewClient(conn.RemoteAddr().String(), conn, m.wsServer)

	m.wsServer.AddClient(client)

	go client.Read()

}

var (
	handleFunc = &MyWebhookHandler{
		wsServer: ws.NewServer(),
	}
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
