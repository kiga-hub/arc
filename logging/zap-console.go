package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewFileCore New is similar to Config.Build except that info and error logs are separated
// only json/console encoder is supported (zap doesn't provide a way to refer to other encoders)
func newConsoleCore(cfg zap.Config) ([]zapcore.Core, error) {
	sink, errSink, err := openSinks(cfg)
	if err != nil {
		return nil, err
	}

	encoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
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

func newConsoleProdCore(level zap.AtomicLevel) ([]zapcore.Core, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = encodeTime
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	var outputPaths, errorOutputPaths []string
	outputPaths = append(outputPaths, "stdout")
	errorOutputPaths = append(errorOutputPaths, "stderr")
	return newConsoleCore(zap.Config{
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: false,
		// Level:             zap.NewAtomicLevelAt(zapCore.InfoLevel),
		Level:            level,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	})
}

func newConsoleDevCore(level zap.AtomicLevel) ([]zapcore.Core, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = encodeTime
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	var outputPaths, errorOutputPaths []string
	outputPaths = append(outputPaths, "stdout")
	errorOutputPaths = append(errorOutputPaths, "stderr")
	return newConsoleCore(zap.Config{
		Development:       true,
		DisableCaller:     true,
		DisableStacktrace: true,
		// Level:             zap.NewAtomicLevelAt(zapCore.DebugLevel),
		Level:            level,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	})
}
