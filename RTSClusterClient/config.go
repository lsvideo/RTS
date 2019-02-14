// config
package main

import (
	"fmt"

	"github.com/astaxie/beego/config/yaml"
)

type SL_config struct {
	IP         string
	Port       int
	Type       string
	ZK_servers []string
}

type ServerType int32

const (
	ServerTypeUnknown        ServerType = 0
	ServerTypeClusterManager ServerType = 1
	ServerTypeRTMP           ServerType = 2
)

var (
	serverTypeNames = map[ServerType]string{
		ServerTypeUnknown:        "ServerTypeUnknown",
		ServerTypeClusterManager: "ServerTypeClusterManager",
		ServerTypeRTMP:           "ServerTypeRTMP",
	}
)

func GetServerType(stype string) ServerType {
	switch stype {
	case "rtmp", serverTypeNames[ServerTypeRTMP]:
		return ServerTypeRTMP
	case "clustermanager", serverTypeNames[ServerTypeClusterManager]:
		return ServerTypeClusterManager
	default:
		return ServerTypeUnknown
	}
}

func Parse_config(cfgfile string, sl_cfg *SL_config) error {
	conf, err := yaml.ReadYmlReader(cfgfile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sl_cfg.IP = conf["ip"].(string)
	sl_cfg.Port = int(conf["port"].(int64))
	sl_cfg.Type = conf["type"].(string)
	servers := conf["zk_servers"]

	switch v := servers.(type) {
	case []interface{}:
		//fmt.Println("zk_servers is an array: ", len(v))
		sl_cfg.ZK_servers = make([]string, len(v))
		for i, u := range v {
			fmt.Println(i, u)
			sl_cfg.ZK_servers[i] = u.(string)
		}
	}

	//fmt.Println(sl_cfg)
	return nil
}
