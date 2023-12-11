package logging

import (
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"common/logging/conf"
	microConf "common/micro/conf"
)

// NewFileCore New is similar to Config.Build except that info and error logs are separated
// only json/console encoder is supported (zap doesn't provide a way to refer to other encoders)
func newFileCore(cfg zap.Config) ([]zapcore.Core, error) {
	sink, errSink, err := openSinks(cfg)
	if err != nil {
		return nil, err
	}

	encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
	logLevel := cfg.Level.Level()
	stdoutPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= logLevel && lvl < zapcore.ErrorLevel
	})
	stderrPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	return []zapcore.Core{
		zapcore.NewCore(encoder, sink, stdoutPriority),
		zapcore.NewCore(encoder, errSink, stderrPriority),
	}, nil
}

func openSinks(cfg zap.Config) (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	sink, closeOut, err := zap.Open(cfg.OutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err := zap.Open(cfg.ErrorOutputPaths...)
	if err != nil {
		closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}

// 格式化时间
func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newFileProdCore(basic *microConf.BasicConfig, config *conf.LogConfig, level zap.AtomicLevel) ([]zapcore.Core, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = encodeTime
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	var outputPaths, errorOutputPaths []string
	outputPaths = append(outputPaths, filepath.Join(config.Path, basic.Service)+"."+basic.Instance+".prod.out.log")
	errorOutputPaths = append(errorOutputPaths, filepath.Join(config.Path, basic.Service)+"."+basic.Instance+".prod.error.log")
	return newFileCore(zap.Config{
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: false,
		// Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Level:            level,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	})
}

func newFileDevCore(basic *microConf.BasicConfig, config *conf.LogConfig, level zap.AtomicLevel) ([]zapcore.Core, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = encodeTime
	var outputPaths, errorOutputPaths []string
	outputPaths = append(outputPaths, filepath.Join(config.Path, basic.Service)+"."+basic.Instance+".dev.out.log")
	errorOutputPaths = append(errorOutputPaths, filepath.Join(config.Path, basic.Service)+"."+basic.Instance+".dev.error.log")
	return newFileCore(zap.Config{
		Development:       true,
		DisableCaller:     true,
		DisableStacktrace: true,
		// Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Level:            level,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	})
}
