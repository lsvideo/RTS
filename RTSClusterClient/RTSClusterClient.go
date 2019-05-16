package main

import (
	"os"
	"os/signal"

	"sl_log"
	//"strconv"
	"syscall"

	//"time"
	"zkhelper"
)

var config SL_config
var APP_NAME = GetAppName()
var log = sl_log.Log
var serverdone = make(chan int, 1)
var signalChan = make(chan os.Signal, 1)
var serverexit = make(chan int, 1)

var rtsclient zkhelper.ZKClient

func init() {
	sl_log.SetLogLevel("info")
	sl_log.SetLogPath("./" + GetAppName() + ".log")
	Parse_config("./sl.yaml", &config)
	//Parse_config("./cm.yaml", &config)
	log.Infoln("get config:", config)

}

func exitclean() int {
	signal.Notify(signalChan, os.Interrupt, os.Kill) //进程采集信号量。
	select {
	case <-signalChan:
		log.Infoln("signal exit")
		break
	case <-serverexit:
		log.Infoln("server exit")
		break
	}

	st := GetServerType(config.Type)
	service := mapService[st]
	service.SLServiceStop()
	log.Infoln("exit")
	return 0
}

func main() {
	defer PanicRecover()()
	st := GetServerType(config.Type)
	service := mapService[st]
	go service.SLServiceStart()

	logFile, err := os.OpenFile("./fatal.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		log.Println("服务启动出错", "打开异常日志文件失败", err)
		return
	}
	// 将进程标准出错重定向至文件，进程崩溃时运行时将向该文件记录协程调用栈信息
	syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))

	exitclean()

}
