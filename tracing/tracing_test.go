package tracing

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/kiga-hub/arc/logging"
	logConf "github.com/kiga-hub/arc/logging/conf"
	"github.com/kiga-hub/arc/logging/loki"
	microConf "github.com/kiga-hub/arc/micro/conf"
)

func TestTraceLoggingToLoki(t *testing.T) {
	logConfig := logConf.GetLogConfig()
	//LokiAddr: "192.168.26.233:9096",

	traceConfig := GetTraceConfig()
	//JaegerCollectorAddr: "http://192.168.26.233:14268",

	config := microConf.BasicConfig{
		Zone:       "default",
		Machine:    "one",
		Service:    "test-service",
		Instance:   "local",
		AppVersion: "v1.0.0",
		AppName:    "test_loki",
	}
	logConfig.Level = "ERROR"
	logger, err := logging.CreateLogger(&config, logConfig)
	if err != nil {
		t.Logf("%v", err)
		return
	}

	tracer, close, err := CreateTracer(config, *traceConfig, logger)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer close.Close()
	span := tracer.StartSpan("main")
	traceID := GetTraceIDFromSpan(span)
	taskID := "task-" + traceID
	fmt.Println(traceID)
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	_, finish, zlogger := StartChildSpan(ctx, logger.Sugar(), tracer, "testspan-1")
	zlogger.Infow("span test info", "model", "tester", "taskID", taskID)
	zlogger.Warn("span test warn", "arg2", "arg1")
	zlogger.Errorf("span test error %s %s", "model", "tester")
	finish()

	ctx, finish1, _ := StartChildSpan(ctx, logger.Sugar(), tracer, "testspan-2")
	_, finish2, _ := StartChildSpan(ctx, logger.Sugar(), tracer, "testspan-2-1")
	finish2()
	finish1()

	span.Finish()

	fmt.Println("wait 3 seconds")
	time.Sleep(time.Second * 3)

	lokiClient, err := loki.NewLokiClient(logConfig.LokiAddr) //"192.168.9.9:9096"
	if err != nil {
		t.Logf("%v", err)
		return
	}
	client, err := NewTempoClient(traceConfig.TempoQueryAddr) //"192.168.9.9:9095"
	if err != nil {
		t.Logf("%v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	traceIDs, err := GetTraceIDFromLog(ctx, lokiClient, fmt.Sprintf(`{taskID="%s"}`, taskID))
	if err != nil {
		t.Errorf("%v", err)
	}
	fmt.Println(traceIDs)
	spans, err := client.FindTraceByID(ctx, traceIDs[0])
	if err != nil {
		t.Errorf("%v", err)
	}
	for k, v := range spans {
		fmt.Printf("=========== %d ============\n", k)
		for _, vv := range v.Resource.Attributes {
			fmt.Printf("%s = %s\n", vv.Key, vv.Value.GetStringValue())
		}
		for _, vv := range v.InstrumentationLibrarySpans {
			for _, vvv := range vv.Spans {
				spanID := hex.EncodeToString(vvv.SpanId)
				fmt.Printf("%s|%s: %d - %d\n", spanID, vvv.Name, vvv.StartTimeUnixNano, vvv.EndTimeUnixNano)
			}
		}
	}
}
