package main

import (
	"os"
	"os/signal"

	"sl_log"
	//"strconv"
	//"time"
)

var (
	REPORT_INTERVAL = 1
)

var config SL_config
var APP_NAME = GetAppName()
var log = sl_log.Log
var serverdone = make(chan int, 1)
var signalChan = make(chan os.Signal, 1)
var serverexit = make(chan int, 1)

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

	switch GetServerType(config.Type) {
	case ServerTypeRTMP:
		serverdone <- 1
		<-serverdone
		break
	}
	log.Infoln("!!!!!!!!!!!!!! exit")
	return 0
}

func main() {

	Parse_config("./sl.yaml", &config)
	//Parse_config("./cm.yaml", &config)

	log.Infoln("get config:", config)
	sl_log.SetLogLevel("debug")

	switch GetServerType(config.Type) {
	case ServerTypeRTMP:
		go videoserverstart()
	case ServerTypeClusterManager:
		go clustermanagerstart()
		go taskmanager()
	default:
		log.Errorf("%s:%s unknown service type!", APP_NAME, GetFuncName())
		serverexit <- 1
	}

	exitclean()

}
