package error

import (
	"errors"
	"fmt"
	"time"
)

// FormatError 格式化打印错误
//
//goland:noinspection GoUnusedExportedFunction
func FormatError(who string, err error) string {
	return fmt.Sprintf("%s %s %s", who, time.Now().Format("2006-01-02 15:04:05.999"), err.Error())
}

// ErrNoServiceAvailable 没有可用的服务
var ErrNoServiceAvailable = errors.New("no services available")

// ErrNoTaskInTaskResult 在任务结果中勿失任务
var ErrNoTaskInTaskResult = errors.New("TaskID not found in task_result")

// ErrFileProperty 文件属性错误
var ErrFileProperty = errors.New("file property error")

// ErrUnavailableHost 不可用的主机
var ErrUnavailableHost = errors.New("unavailable host")

// ErrPoolIsClosed 线程池是关闭的
var ErrPoolIsClosed = errors.New("pool is closed")

// ErrInvalidCapacitySettings 无效的容量设置
var ErrInvalidCapacitySettings = errors.New("invalid capacity settings")

// ErrConnectionIsNil 连接是空。拒绝
var ErrConnectionIsNil = errors.New("connection is nil. rejecting")

// ErrConfigNotHaveConfig 配置文件err:没有配置env
var ErrConfigNotHaveConfig = errors.New("config file err: don't have the config env")

// ErrRedisFail redis失败
var ErrRedisFail = errors.New("redis fail")

// ErrLockFail 锁失败
var ErrLockFail = errors.New("lock fail")

// ErrDbConnection 数据库连接错误
var ErrDbConnection = errors.New("DB Connection error")

// ErrTaskNotCompleted 任务没有完成
var ErrTaskNotCompleted = errors.New("the task is not completed")

// ErrNoAdapterServer 没有适配器服务器
var ErrNoAdapterServer = errors.New("no adapter server")

// ErrUnknownType 未知类型
var ErrUnknownType = errors.New("unknown type")

// ErrWorkerDidNotReturn 工作没有返回
var ErrWorkerDidNotReturn = errors.New("worker did not return")

// ErrClientIDInvalid 客户端编号无效
var ErrClientIDInvalid = errors.New("client ID Invalid")

// ErrTaskTypeInvalid 任务任务类型无效
var ErrTaskTypeInvalid = errors.New("task Type Invalid")

// ErrParams 参数错误
var ErrParams = errors.New("params error")

// ErrAlreadyExists 已经退出
var ErrAlreadyExists = errors.New("already exists")
