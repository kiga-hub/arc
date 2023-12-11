package error

import (
	"errors"
	"fmt"
	"time"
)

//FormatError 格式化打印错误
func FormatError(who string, err error) string {
	return fmt.Sprintf("%s %s %s", who, time.Now().Format("2006-01-02 15:04:05.999"), err.Error())
}

//ErrNoServiceAvailable 没有可用的服务
var ErrNoServiceAvailable error = errors.New("No services available")

//ErrNoTaskInTaskResult 在任务结果中勿失任务
var ErrNoTaskInTaskResult error = errors.New("TaskID not found in task_result")

//ErrTextgridMerge TextGrid合并错误
var ErrTextgridMerge error = errors.New("TextGrid Merge error")

//ErrFileProperty 文件属性错误
var ErrFileProperty error = errors.New("File property error")

//ErrUnavailableHost 不可用的主机
var ErrUnavailableHost error = errors.New("unavailable host")

//ErrPoolIsClosed 线程池是关闭的
var ErrPoolIsClosed error = errors.New("pool is closed")

//ErrInvalidCapacitySettings 无效的容量设置
var ErrInvalidCapacitySettings error = errors.New("invalid capacity settings")

//ErrConnectionIsNil 连接是空。拒绝
var ErrConnectionIsNil error = errors.New("connection is nil. rejecting")

//ErrConfigNotHaveConfig 配置文件err:没有配置env
var ErrConfigNotHaveConfig error = errors.New("config file err: don't have the config env")

//ErrRedisFail redis失败
var ErrRedisFail error = errors.New("redis fail")

//ErrLockFail 锁失败
var ErrLockFail error = errors.New("lock fail")

//ErrDbConnection 数据库连接错误
var ErrDbConnection error = errors.New("DB Connection error")

//ErrTaskNoComplished 任务没有完成
var ErrTaskNoComplished error = errors.New("The task is not completed")

//ErrNoAdapterServer 没有适配器服务器
var ErrNoAdapterServer error = errors.New("No adapter server")

//ErrUnknownType 未知类型
var ErrUnknownType error = errors.New("Unknown type")

//ErrWorkerDidNotReturn 工作没有返回
var ErrWorkerDidNotReturn error = errors.New("Worker did not return")

//ErrCuidInvalid 人员编号无效
var ErrCuidInvalid error = errors.New("PersonID Invalid")

//ErrClientIDInvalid 客户端编号无效
var ErrClientIDInvalid error = errors.New("Client ID Invalid")

//ErrTaskTypeInvalid 任务任务类型无效
var ErrTaskTypeInvalid error = errors.New("Task Type Invalid")

//ErrInvalidAudioFilePath 无效的音频文件路径
var ErrInvalidAudioFilePath error = errors.New("Invalid audio file path")

//ErrRegisterAudioFileListEmpty 注册音频文件列表为空
var ErrRegisterAudioFileListEmpty error = errors.New("Registration audio file list is empty")

//ErrInvalidCuidFilePath 无效的cuid文件路径
var ErrInvalidCuidFilePath error = errors.New("Invalid cuid file path")

//ErrParams 参数错误
var ErrParams = errors.New("params error")

//ErrAlreadyExists 已经退出
var ErrAlreadyExists = errors.New("already exists")
