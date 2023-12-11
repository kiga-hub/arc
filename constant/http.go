package constant

const (
	//UUIDLength uuid长度
	UUIDLength int = 36
	//URLMetrics metrics的地址
	URLMetrics string = "/metrics"
	//EchoLogFormat echo日志格式
	EchoLogFormat string = "${time_rfc3339_nano} method=${method}, uri=${uri}, status=${status}\n"
)
