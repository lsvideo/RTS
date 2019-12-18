package main

import (
	"os"
	"os/signal"

	"sl_log"
	//"strconv"
	"flag"
	"fmt"
	"syscall"

	//"time"
	"RTSClusterClient/version"
	"zkhelper"
)

var config SL_config
var APP_NAME = GetAppName()
var log = sl_log.Log
var serverdone = make(chan int, 1)
var signalChan = make(chan os.Signal, 1)
var serverexit = make(chan int, 1)

var rtsclient zkhelper.ZKClient
var Version string

func init() {
	var showVersion bool = false
	var cfgFile string = ""
	flag.BoolVar(&showVersion, "v", false, "current version")
	flag.BoolVar(&showVersion, "version", false, "current version")
	flag.StringVar(&cfgFile, "c", "", "config file")
	flag.StringVar(&cfgFile, "config", "", "config file")
	flag.Parse()

	if showVersion == true {
		fmt.Println(version.FullVersion())
		os.Exit(0)
	}
	if cfgFile == "" {
		fmt.Println("Please specify configuration file!")
		os.Exit(0)
	}

	sl_log.SetLogLevel("debug")
	sl_log.SetLogPath("./" + GetAppName() + ".log")
	config.Parse_config(cfgFile)
	log.Debugln("get config:", config)

	initMysql(config.DB_server, config.DB_port, config.DB_user, config.DB_pwd, config.DB_name)

}

func exitclean() int {

	logFile, err := os.OpenFile("./fatal.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		log.Println("服务启动出错", "打开异常日志文件fatal.log失败", err)
		return 1
	}
	// 将进程标准出错重定向至文件，进程崩溃时运行时将向该文件记录协程调用栈信息
	syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))

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
	go cronStart()
	defer cronStop()
	exitclean()
}
