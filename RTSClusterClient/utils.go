// utils
package main

import (
	"fmt"
	//"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	gopsutil_net "github.com/shirou/gopsutil/net"
)

type MemStatus struct {
	All         uint64  `json:"all"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Self        uint64  `json:"self"`
	Pagesize    uint64  `json:"pagesize"`
	UsedPercent float64 `json:"usedpercent"`
}

type SysState struct {
	Cpu   float64 `json:"cpu"`
	Mem   float64 `json:"mem"`
	Links int     `json:"links"`
	NetRX uint64  `json:"netrx"`
	NetTX uint64  `json:"nettx"`
}

func Check_err(e error) {
	if e != nil {
		panic(e)
	}
}

func CPUnumbers() int {
	return runtime.NumCPU()
}

func Decimal(value float64) float64 {
	//return math.Trunc(value*1e2+0.5) * 1e-2
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

func MemStat() MemStatus {
	//自身占用
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	mem := MemStatus{}
	mem.Self = memStat.Alloc

	//系统占用,仅linux/mac下有效
	//system memory usage
	mem.Pagesize = uint64(syscall.Getpagesize())
	sysInfo := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysInfo)
	if err == nil {
		mem.All = sysInfo.Totalram
		mem.Free = sysInfo.Freeram
		mem.Used = mem.All - mem.Free
		mem.UsedPercent = float64(mem.Used) / float64(mem.All) * 100.0
	}
	return mem
}

func MemUsedPersent() float64 {
	v, _ := mem.VirtualMemory()
	return Decimal(v.UsedPercent)
}

func CpuUsedPersent() float64 {
	persent, _ := cpu.Percent(0, false)
	return Decimal(persent[0])
}

func BandwidthUsed(dev string) (uint64, uint64) {
	netioall, _ := gopsutil_net.IOCounters(true)
	for _, netio := range netioall {
		if netio.Name == dev {
			return netio.BytesSent, netio.BytesRecv
		}
	}
	//fmt.Println(iostat)
	return 0, 0
}

func GetInterfaceNameFromIp(ip string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("%s:%s Can not get local interface info:%s", APP_NAME, GetFuncName(), err)
		os.Exit(-1)
	}
	for _, inter := range interfaces {
		flags := inter.Flags.String()
		if strings.Contains(flags, "up") && strings.Contains(flags, "broadcast") {
			//mac_address := inter.HardwareAddr.String()

			addrs, _ := inter.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						//fmt.Println(ipnet.IP.String())
						if ip == ipnet.IP.String() {
							//fmt.Println(inter.Name)
							return inter.Name
						}
					}
				}
			}

		}
	}
	return ""
}

func GetAppName() string {
	cmd := string(os.Args[0])
	lastIndex := strings.LastIndex(cmd, "/")
	appname := string(cmd[lastIndex+1:])
	return appname
}

func GetFuncName() string {
	//pc, file, line, ok := runtime.Caller(1)
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		func_name := runtime.FuncForPC(pc).Name()
		//fmt.Println("file:", file, " func_name", func_name, " line:", line)
		return func_name
	}
	return ""
}

func Func_runningtime_trace() func() {
	start := time.Now()
	pc, _, _, _ := runtime.Caller(1)
	func_name := runtime.FuncForPC(pc).Name()

	fmt.Printf("%s %s start\n", time.Now().Format("2006-01-02 15:04:05"), func_name)
	return func() {
		fmt.Printf("%s %s exit (%s)\n", time.Now().Format("2006-01-02 15:04:05"), func_name, time.Since(start))
	}
}

func Between(str, starting, ending string) string {
	s := strings.Index(str, starting)
	if s < 0 {
		return ""
	}
	s += len(starting)
	e := strings.Index(str[s:], ending)
	if e < 0 {
		return ""
	}
	return str[s : s+e]
}
