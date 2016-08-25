package logs

import (
	"io"
	"log"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLog(debugHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer, flag int) {
	Debug = log.New(debugHandle, "[INTOOLS] [DEBUG] ", flag)
	Info = log.New(infoHandle, "[INTOOLS] [INFO]  ", flag)
	Warning = log.New(warningHandle, "[INTOOLS] [WARN]  ", flag)
	Error = log.New(errorHandle, "[INTOOLS] [ERROR] ", flag)
}
