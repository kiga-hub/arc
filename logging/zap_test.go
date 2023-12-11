package logging

import (
	"fmt"
	"testing"
	"time"

	logConf "common/logging/conf"
	microConf "common/micro/conf"
)

func TestGraylog(t *testing.T) {
	logConfig := logConf.GetLogConfig()
	logger, err := CreateLogger(&microConf.BasicConfig{
		Zone:       "Prod1",
		IsDevMode:  true,
		AppVersion: "v1.0.0",
		AppName:    "test_graylog",
	}, logConfig)
	if err != nil {
		t.Errorf("%v", err)
	}
	l := logger.Sugar()
	l.Debugw("zap err test", "model", "test")
	l.Infow("zap info test", "model", "test")
	l.Errorw("zap err test", "model", "test")
}

func TestLoki(t *testing.T) {
	logConfig := logConf.GetLogConfig()
	logger, err := CreateLogger(&microConf.BasicConfig{
		Zone:       "Prod1",
		IsDevMode:  true,
		AppVersion: "v1.0.0",
		AppName:    "test_loki",
	}, logConfig)
	if err != nil {
		t.Logf("%v", err)
		return
	}
	l := logger.Sugar()
	l.Debugw("zap err test", "model", "modelA", "number", 99)
	l.Infow("traceID=f4b0ce7044ada1e2 | zap info test ", "model", "modelB", "float", 12.2345)
	l.Errorw("zap err test", "model", "modelC", "number", 100)

	fmt.Println("wait 5 seconds")
	time.Sleep(time.Second * 5)
}
