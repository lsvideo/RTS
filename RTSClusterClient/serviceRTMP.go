// videoserver
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	//"syscall"

	"zkhelper"
)

var (
	REPORT_INTERVAL = 1
)

// RTMP服务
type ServiceRTMP struct {
}

var rtmp ServiceRTMP

func init() {
	log.Println("init rtmp")
	mapService[ServerTypeRTMP] = rtmp
}

func (rtmp ServiceRTMP) SLServiceStart() {
	//func videoserverstart() {
	//defer Func_runningtime_trace()()
	defer PanicRecover()()
	servers := config.ZK_servers
	_, err := rtsclient.NewClient(servers, config.Type, config.IP, config.Port, 1)
	Check_err(err)

	defer rtsclient.Close()

	var autonode zkhelper.ZKNode
	autonode.ServiceType = zkhelper.GetServiceType(config.Type)
	autonode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(autonode.ServiceType), zkhelper.NodeTypeAutoDetected)
	autonode.Name = "[" + config.IP + ":" + strconv.Itoa(config.Port) + "]"
	rtsclient.Register(&autonode)

	eChatCannelInit()

	go srsmanager()

	//bandwidth used
	var sysstate SysState
	var wblast, rblast uint64 = 0, 0
	for {
		//report info
		mem := MemUsedPersent()
		cpu := CpuUsedPersent()
		//wb, rb := BandwidthUsed(GetInterfaceNameFromIp(config.IP))
		wb, rb := BandwidthUsed(config.Device)
		sum, _ := get_summaries("127.0.0.1:" + strconv.Itoa(config.Srs_api_port))
		if sum != nil {
			if sysstate.Links == 0 {
				log.Errorln("srs start!")
			}
			sysstate.Links = sum.Date.System.Conn_srs
		} else {
			log.Errorln("srs do not start!")
			sysstate.Links = 0
		}
		wb -= wblast
		rb -= rblast
		if wblast != 0 && rblast != 0 {
			os.Stdout.WriteString(fmt.Sprintf("\rCPU: %.2f    MEM:  %.2f   Links:%d   Down:%d KB/s    Up:%d KB/s    ", cpu, mem, sysstate.Links, rb/1024/(uint64(REPORT_INTERVAL)), wb/1024/(uint64(REPORT_INTERVAL))))
			sysstate.Cpu = cpu
			sysstate.Mem = mem
			sysstate.NetRX = rb
			sysstate.NetTX = wb
			sysstate.IP = config.IP
			sysstate.Port = strconv.Itoa(config.Port)
			info, _ := json.Marshal(sysstate)
			//fmt.Println(string(info))
			autonode.Data = info
			rtsclient.Set(&autonode)
		}
		wblast += wb
		rblast += rb

		select {
		case <-serverdone:
			rtsclient.Close()
			fmt.Println("exiting...")
			serverdone <- 1
			break
		case <-time.After(time.Duration(REPORT_INTERVAL) * time.Second):
			break
		}
	}

}

func (rtmp ServiceRTMP) SLServiceStop() {
	serverdone <- 1
	<-serverdone
}
