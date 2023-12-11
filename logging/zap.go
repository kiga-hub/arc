package logging

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"common/logging/conf"
	"common/logging/loki"
	microConf "common/micro/conf"
)

// ILogger log类
type ILogger interface {
	Debugw(string, ...interface{})
	Infow(string, ...interface{})
	Warnw(string, ...interface{})
	Errorw(string, ...interface{})
	Panicw(string, ...interface{})
	Fatalw(string, ...interface{})

	Panic(args ...interface{})
	Fatal(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Panicf(string, ...interface{})
}

//Logger  log全局对象
//var Logger ILogger

// CreateLogger 创建一个 log对象
func CreateLogger(basicConfig *microConf.BasicConfig, logConfig *conf.LogConfig) (*zap.Logger, error) { // logger, error
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(logConfig.Level))
	if err != nil {
		return nil, fmt.Errorf("Fatal to parse logger level: %s", err)
	}

	cores := []zapcore.Core{}

	// graylog -
	if logConfig.GraylogAddr != "" {
		core, err := newGraylogCore(logConfig)
		if err != nil {
			return nil, err
		}
		cores = append(cores, core)
	}

	// loki -
	if logConfig.LokiAddr != "" {
		core, err := loki.NewLokiCore(logConfig)
		if err != nil {
			return nil, err
		}
		cores = append(cores, core)
	}

	// file -
	if logConfig.Path != "" {
		if basicConfig.IsDevMode {
			fcores, err := newFileDevCore(basicConfig, logConfig, level)
			if err != nil {
				return nil, err
			}
			cores = append(cores, fcores...)
		} else {
			fcores, err := newFileProdCore(basicConfig, logConfig, level)
			if err != nil {
				return nil, err
			}
			cores = append(cores, fcores...)
		}
	}

	// console -
	if logConfig.Console {
		if basicConfig.IsDevMode {
			ccores, err := newConsoleDevCore(level)
			if err != nil {
				return nil, err
			}
			cores = append(cores, ccores...)
		} else {
			ccores, err := newConsoleProdCore(level)
			if err != nil {
				return nil, err
			}
			cores = append(cores, ccores...)
		}
	}

	logFields := []zapcore.Field{}
	for _, key := range strings.Split(logConfig.Fields, ",") {
		switch key {
		case "zone":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.Zone})
		case "node":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.Node})
		case "machine":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.Machine})
		case "instance":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.Instance})
		case "service":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.Service})
		case "appname":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.AppName})
		case "appversion":
			logFields = append(logFields, zapcore.Field{Key: key, Type: zapcore.StringType, String: basicConfig.AppVersion})
		}
	}

	tee := zapcore.NewTee(cores...)
	zlog := zap.New(
		tee,
		zap.AddCaller(),
		zap.AddStacktrace(zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= zapcore.WarnLevel
		})),
		zap.Fields(logFields...),
	)
	return zlog, err
}

// LoggerGroup is a group of logger, which provide loggers w/ different levels for modules
type LoggerGroup struct {
	moduleLevel   *sync.Map
	levelLogger   *sync.Map
	defaultLogger ILogger
	defaultLevel  string
}

const (
	// LevelDebug logs are typically voluminous, and are usually disabled in production.
	LevelDebug = "DEBUG"
	// LevelInfo is the default logging priority.
	LevelInfo = "INFO"
	// LevelWarn logs are more important than Info, but don't need individual human review.
	LevelWarn = "WARN"
	// LevelError logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	LevelError = "ERROR"
	// LevelDPanic  logs are particularly important errors. In development the
	// logger panics after writing the message.
	LevelDPanic = "DPANIC"
	// LevelPanic logs a message, then panics.
	LevelPanic = "PANIC"
	// LevelFatal  logs a message, then calls os.Exit(1).
	LevelFatal = "FATAL"
)

var allLevels = []string{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelDPanic, LevelPanic, LevelFatal}

// CreateLoggerGroup create a logger group
func CreateLoggerGroup(basicConfig *microConf.BasicConfig, logConfig *conf.LogGroupConfig) (*LoggerGroup, error) {
	levelLogger := sync.Map{}

	if basicConfig != nil && logConfig != nil {
		for _, v := range allLevels {
			logger, err := CreateLogger(basicConfig, &conf.LogConfig{
				Level:       v,
				Path:        logConfig.Path,
				LokiAddr:    logConfig.LokiAddr,
				GraylogAddr: logConfig.GraylogAddr,
				Console:     logConfig.Console,
			})
			if err != nil {
				return nil, err
			}
			levelLogger.Store(strings.ToUpper(v), logger.Sugar())
		}
	}
	group := &LoggerGroup{
		moduleLevel:   &sync.Map{},
		levelLogger:   &levelLogger,
		defaultLogger: &NoopLogger{},
		defaultLevel:  LevelInfo,
	}

	if logConfig != nil {
		group.defaultLevel = logConfig.Level
		l, ok := levelLogger.Load(strings.ToUpper(logConfig.Level))
		if ok {
			group.defaultLogger = l.(*zap.SugaredLogger)
		}
		for l, modules := range logConfig.Levels {
			for _, module := range modules {
				group.SetLevel(module, l)
			}
		}
	}
	return group, nil
}

// M return the logger for the module, will return logger w/ INFO level if level is not set for the module
func (l *LoggerGroup) M(module string) ILogger {
	level, ok := l.moduleLevel.Load(strings.ToUpper(module))
	if !ok {
		return l.defaultLogger
	}
	logger, ok := l.levelLogger.Load(level)
	if !ok {
		return l.defaultLogger
	}
	return logger.(*zap.SugaredLogger)
}

// SetLevel set log level for the module, which can be called in runtime
func (l *LoggerGroup) SetLevel(module, level string) {
	if level != LevelDebug &&
		level != LevelInfo &&
		level != LevelWarn &&
		level != LevelError &&
		level != LevelDPanic &&
		level != LevelPanic &&
		level != LevelFatal {
		_, ok := l.moduleLevel.Load(strings.ToUpper(module))
		if ok {
			return
		}
		level = l.defaultLevel
	}
	l.moduleLevel.Store(strings.ToUpper(module), strings.ToUpper(level))
}

// GetLevels return log levels of each module
func (l *LoggerGroup) GetLevels() map[string]string {
	result := map[string]string{}
	l.moduleLevel.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(string)
		return true
	})
	return result
}

// Sync calls the underlying Core's Sync method
func (l *LoggerGroup) Sync() error {
	var err error
	l.levelLogger.Range(func(key interface{}, value interface{}) bool {
		err = value.(*zap.SugaredLogger).Sync()
		return err == nil
	})
	return err
}
