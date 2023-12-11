package logging

import (
	"os"

	gelf "github.com/snovichkov/zap-gelf"
	"go.uber.org/zap/zapcore"

	"common/logging/conf"
)

func newGraylogCore(config *conf.LogConfig) (zapcore.Core, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return gelf.NewCore(
		gelf.Addr(config.GraylogAddr),
		gelf.Host(host),
		gelf.LevelString(config.Level),
	)
}
