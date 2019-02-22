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

func watchServer(client *zkhelper.ZKClient, server *VServer, status zkhelper.NodeStatus) {
	ch, err := client.WatchNode(&server.statenode, status)
	Check_err(err)
	<-ch
	switch server.statenode.Status {
	case zkhelper.NodeStatusDelete:
		server.statenode.Status = zkhelper.NodeStatusDelete

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
			log.Println(k, v)
			server.statenode.Status = zkhelper.NodeStatusWatched
			mapServers.Store(key, server)
			go watchServer(client, &server, zkhelper.NodeStatusDelete)
		} else if server.statenode.Status == zkhelper.NodeStatusDelete {
			log.Println(k, v)
			CleanServerChannels(server.Name)
			mapServers.Delete(server.Name)
			log.Infoln("delete node", server.Name)
			log.Infoln(mapServers)
			log.Infoln(mapServers.Load(server.Name))
		}

		return true
	}

	for {
		mapServers.Range(watchoneserver)
		time.Sleep(time.Duration(2) * time.Second)
	}
}

func clustermanagerstart() {
	//defer Func_runningtime_trace()()
	servers := config.ZK_servers
	var rtsclient zkhelper.ZKClient
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
	go watchAllServers(&rtsclient)

	var autonode zkhelper.ZKNode //cluster manager服务 自动发现节点
	autonode.ServiceType = zkhelper.GetServiceType(config.Type)
	autonode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(autonode.ServiceType), zkhelper.NodeTypeAutoDetected)
	autonode.Name = "[" + config.IP + ":" + strconv.Itoa(config.Port) + "]"
	rtsclient.Register(&autonode)

	var vnode zkhelper.ZKNode //视频服务自动发现节点
	vnode.NodeType = zkhelper.NodeTypeAutoDetected
	vnode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(zkhelper.ServiceTypeRTMP), zkhelper.NodeTypeAutoDetected)
	for {
		children, err := rtsclient.GetChildren(&vnode)
		Check_err(err)
		for _, path := range children {
			ip_port := Between(path, "[", "]")
			if value, ok := mapServers.Load(ip_port); !ok {
				//不存在
				log.Infoln("New node:", path)
				var server VServer
				server.Name = ip_port
				server.statenode.Name = ip_port
				server.statenode.Status = zkhelper.NodeStatusNew
				server.statenode.Path = vnode.Path + "/" + path
				server.lstChannels = list.New()
				mapServers.Store(ip_port, server)
			} else {
				server := value.(VServer)
				log.Println(server)
			}

		}
		log.Infoln(mapServers)
		rtsclient.WatchChild(&vnode)
		time.Sleep(time.Duration(REPORT_INTERVAL) * time.Second)
	}

}
