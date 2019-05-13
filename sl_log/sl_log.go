// sl_log.go
package sl_log

import (
	//	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger = nil

func init() {
	if Log == nil {
		Log = logrus.New()
		//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
		customFormatter := new(logrus.TextFormatter)
		customFormatter.FullTimestamp = true                        // 显示完整时间
		customFormatter.TimestampFormat = "2006-01-02 15:04:05:001" // 时间格式
		customFormatter.DisableTimestamp = false                    // 禁止显示时间
		customFormatter.DisableColors = true                        // 禁止颜色显示

		Log.SetFormatter(customFormatter)

		//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
		//Log.SetOutput(os.Stdout)
		// You could set this to any `io.Writer` such as a file

		//fmt.Println(os.Hostname())
		//设置最低loglevel
		Log.SetLevel(logrus.InfoLevel)
	}

}

func SetLogLevel(level string) {
	switch level {
	case "info":
		Log.SetLevel(logrus.InfoLevel)
		break
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		Log.SetLevel(logrus.WarnLevel)
		break
	}
}

func SetLogPath(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.Out = file
		Log.Info("Open log file ", path)
	} else {
		Log.Info("Failed to log to file, using default stderr")
		panic(err)
	}
}
