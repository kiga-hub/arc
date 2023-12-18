package error

import (
	"errors"
	"fmt"
	"time"
)

// FormatError FormatError
//
//goland:noinspection GoUnusedExportedFunction
func FormatError(who string, err error) string {
	return fmt.Sprintf("%s %s %s", who, time.Now().Format("2006-01-02 15:04:05.999"), err.Error())
}

// ErrNoServiceAvailable ErrNoServiceAvailable
var ErrNoServiceAvailable = errors.New("no services available")

// ErrNoTaskInTaskResult not found in task_result
var ErrNoTaskInTaskResult = errors.New("TaskID not found in task_result")

// ErrFileProperty file property error
var ErrFileProperty = errors.New("file property error")

// ErrUnavailableHost unavailable host
var ErrUnavailableHost = errors.New("unavailable host")

// ErrPoolIsClosed pool is closed
var ErrPoolIsClosed = errors.New("pool is closed")

// ErrInvalidCapacitySettings invalid capacity settings
var ErrInvalidCapacitySettings = errors.New("invalid capacity settings")

// ErrConnectionIsNil connection is nil. rejecting
var ErrConnectionIsNil = errors.New("connection is nil. rejecting")

// ErrConfigNotHaveConfig config file err: don't have the config env
var ErrConfigNotHaveConfig = errors.New("config file err: don't have the config env")

// ErrRedisFail redis fail
var ErrRedisFail = errors.New("redis fail")

// ErrLockFail lock fail
var ErrLockFail = errors.New("lock fail")

// ErrDbConnection DB Connection error
var ErrDbConnection = errors.New("DB Connection error")

// ErrTaskNotCompleted the task is not completed
var ErrTaskNotCompleted = errors.New("the task is not completed")

// ErrNoAdapterServer no adapter server
var ErrNoAdapterServer = errors.New("no adapter server")

// ErrUnknownType unknown type
var ErrUnknownType = errors.New("unknown type")

// ErrWorkerDidNotReturn worker did not return
var ErrWorkerDidNotReturn = errors.New("worker did not return")

// ErrClientIDInvalid client ID Invalid
var ErrClientIDInvalid = errors.New("client ID Invalid")

// ErrTaskTypeInvalid task Type Invalid
var ErrTaskTypeInvalid = errors.New("task Type Invalid")

// ErrParams params error
var ErrParams = errors.New("params error")

// ErrAlreadyExists already exit
var ErrAlreadyExit = errors.New("already exit")
