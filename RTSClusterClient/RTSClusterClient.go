package main

import (
	"os"
	"os/signal"

	"sl_log"
	//"strconv"

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
	//Parse_config("./sl.yaml", &config)
	sl_log.SetLogLevel("info")
	sl_log.SetLogPath("./" + GetAppName() + ".log")
	Parse_config("./cm.yaml", &config)
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

	st := GetServerType(config.Type)
	service := mapService[st]
	go service.SLServiceStart()

	exitclean()

}
