// clustermanager
package main

import (
	//"fmt"
	"container/list"
	"strconv"
	"sync"
	"time"

	"zkhelper"
)

var mapServers sync.Map
var mapChannels sync.Map

type VServer struct {
	Name        string     `json:"name"` // 名称
	ServerType  ServerType `json:"type"`
	IP          string     `json:"ip"`
	Port        int        `json:"port"`
	ChannelNum  int        `json:"channelnum"`
	state       SysState
	statenode   zkhelper.ZKNode
	datanode    zkhelper.ZKNode
	locknode    zkhelper.ZKNode
	lstChannels *list.List
}

// RTMP服务
type ServiceClusterManager struct {
}

var cm ServiceClusterManager

func init() {
	log.Println("init clustermanager")
	mapService[ServerTypeClusterManager] = cm
}

func watchServer(client *zkhelper.ZKClient, server *VServer, status zkhelper.NodeStatus) {
	ch, err := client.WatchNode(&server.statenode, status)
	Check_err(err)
	<-ch
	switch server.statenode.Status {
	case zkhelper.NodeStatusDelete:
		server.statenode.SetNodeStatus(zkhelper.NodeStatusDelete)
		mapServers.Store(server.Name, *server)

		break
	}
}

func watchAllServers(client *zkhelper.ZKClient) {
	watchoneserver := func(k, v interface{}) bool {
		//这个函数的入参、出参的类型都已经固定，不能修改
		//可以在函数体内编写自己的代码，调用map中的k,v
		server := v.(VServer)
		key := k.(string)
		if server.statenode.Status == zkhelper.NodeStatusNew {
			//log.Println(k, v)
			server.statenode.Status = zkhelper.NodeStatusWatched
			mapServers.Store(key, server)
			go watchServer(client, &server, zkhelper.NodeStatusDelete)
		} else if server.statenode.Status == zkhelper.NodeStatusDelete {
			//log.Println(k, v)
			CleanServerChannels(server.Name)
			log.Infoln("delete node", server.datanode)
			rtsclient.Delete(&server.datanode)
			mapServers.Delete(server.Name)
			//删除data节点

			//log.Infoln(mapServers)
			//log.Infoln(mapServers.Load(server.Name))
		}
		//log.Errorln("every:", server.statenode.Status)
		//log.Errorf("Statenode: %p\n", &server.statenode)
		return true
	}

	for {
		mapServers.Range(watchoneserver)
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (cm ServiceClusterManager) SLServiceStart() {
	//defer Func_runningtime_trace()()
	servers := config.ZK_servers
	_, err := rtsclient.NewClient(servers, config.Type, config.IP, config.Port, 3)
	Check_err(err)

	defer rtsclient.Close()
	/*
		var locknode zkhelper.ZKNode
		locknode.Type = zkhelper.GetNodeType(config.Type)
		locknode.Name = "[" + config.IP + ":" + strconv.Itoa(config.Port) + "]"
		locknode.Path = "/shanli/lock/" + locknode.Name
		log.Println("lock!!!!!!!!!!")
		err, lockch := rtsclient.Lock(&locknode, zkhelper.LockModeWrite)
		node := <-lockch
		log.Println(node)
		time.Sleep(time.Duration(10) * time.Second)
		log.Println("Do something!!!!!!!!!!")
		rtsclient.UnLock(&locknode)
		log.Println("unlock!!!!!!!!!!")

		//lock test

	*/
	go taskmanager()

	go watchAllServers(&rtsclient)

	var autonode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	autonode.SetServiceType(zkhelper.GetServiceType(config.Type))
	autonode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(autonode.ServiceType), zkhelper.NodeTypeAutoDetected))
	autonode.SetName("[" + config.IP + ":" + strconv.Itoa(config.Port) + "]")
	log.Println(autonode)
	rtsclient.Register(&autonode)

	var vnode zkhelper.ZKNode //监控rtmp服务自动发现节点
	vnode.SetNodeType(zkhelper.NodeTypeAutoDetected)
	vnode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(zkhelper.ServiceTypeRTMP), zkhelper.NodeTypeAutoDetected))
	log.Println(vnode)
	for {
		children, err := rtsclient.GetChildren(&vnode)
		Check_err(err)
		for _, path := range children {
			ip_port := Between(path, "[", "]")
			//if value, ok := mapServers.Load(ip_port); !ok {
			if _, ok := mapServers.Load(ip_port); !ok {
				//不存在
				log.Infoln("New node:", path)
				var server VServer
				server.SetName(ip_port)
				server.SetServerType(ServerTypeRTMP)
				server.NodeInit(path)
				server.lstChannels = list.New()
				mapServers.Store(ip_port, server)
				//添加server zookeeper路径
				rtsclient.Create(&server.datanode)
				//			} else {
				//				server := value.(VServer)
			}

		}
		rtsclient.WatchChild(&vnode)
	}

}

func (cm ServiceClusterManager) SLServiceStop() {

}

func (server *VServer) NodeInit(autopath string) {
	ip_port := Between(autopath, "[", "]")
	server.statenode.Name = ip_port
	server.statenode.Status = zkhelper.NodeStatusNew
	server.statenode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(zkhelper.ServiceTypeRTMP), zkhelper.NodeTypeAutoDetected) + "/" + autopath

	server.datanode.Name = ip_port
	server.datanode.Status = zkhelper.NodeStatusNew
	server.datanode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(zkhelper.ServiceTypeRTMP), zkhelper.NodeTypeServer) + "/" + ip_port
	server.datanode.Date = []byte("1")
}

func (server *VServer) SetName(name string) {
	server.Name = name
}

func (server *VServer) SetServerType(st ServerType) {
	server.ServerType = st
}
