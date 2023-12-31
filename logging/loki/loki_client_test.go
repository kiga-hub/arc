package loki

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kiga-hub/arc/logging/conf"
)

func TestLokiClient_GetLogs(t *testing.T) {
	logConfig := conf.GetLogConfig()
	client, err := NewLokiClient(logConfig.LokiAddr) //"127.0.0.1:9096"
	if err != nil {
		t.Logf("%v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := client.GetLogs(ctx, `{taskID="123"}`, time.Unix(0, 0), time.Now().UTC(), 1)
	if err != nil {
		t.Errorf("%v", err)
	}
	//spew.Dump(resp)
	if len(resp) >= 0 {
		for k, v := range resp[0].LabelSet {
			fmt.Printf("%s, %s\n", k, v)
		}
	}
}
