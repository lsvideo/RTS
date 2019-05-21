// utils
package main

import (
	"fmt"
	//"math"
	//"bytes"
	"crypto/md5"
	"encoding/hex"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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
	IP    string  `json:"ip"`
	Port  string  `json:"port"`
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

func MapGoThrough(k, v interface{}) bool {
	log.Warningln(k, ":", v)
	return true
}

func GetStringMd5(s string) string {
	md5 := md5.New()
	md5.Write([]byte(s))
	md5Str := hex.EncodeToString(md5.Sum(nil))
	return md5Str
}

func PanicRecover() func() {
	return func() {
		if r := recover(); r != nil {
			log.Errorf("捕获到的错误：%s\n", r)
		}
	}
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func FileCreateTime(filename string) {
	finfo, _ := os.Stat(filename)
	// Sys()返回的是interface{}，所以需要类型断言，不同平台需要的类型不一样，linux上为*syscall.Stat_t
	stat_t := finfo.Sys().(*syscall.Stat_t)
	fmt.Println(stat_t)
	// atime，ctime，mtime分别是访问时间，创建时间和修改时间，具体参见man 2 stat
	fmt.Println(timespecToTime(stat_t.Atim))
	fmt.Println(timespecToTime(stat_t.Ctim))
	fmt.Println(timespecToTime(stat_t.Mtim))
}

func GetFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

func MoveFile(oldPath string, newPath string) error {
	err := os.Rename(oldPath, newPath)
	return err
}

func GetVideoDuration(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func Exec_shell(s string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	//cmd := exec.Command("/bin/bash", "-c", s)
	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	//err := cmd.Run()
	//log.Errorln("Cmd :", s, " Err:", err)
	//return out.String(), err

	cmd := exec.Command("/bin/bash", "-c", s)
	out, err := cmd.CombinedOutput()
	log.Debugln("Cmd :", s, " Err:", err)
	str := string(out)
	str = strings.Replace(str, "\n", "", -1)
	return str, err
}
