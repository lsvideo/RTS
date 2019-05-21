// config
package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type SL_config struct {
	IP               string
	Port             int
	Type             string
	Device           string
	ZK_servers       []string
	Penetrate_server string
	Video_server     string
	Dvr_path         string
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
	yamlFile, err := ioutil.ReadFile(cfgfile)
	log.Println("yamlFile:", yamlFile)
	if err != nil {
		log.Printf("yamlFile.Get err #%v \n", err)
	}
	err = yaml.Unmarshal(yamlFile, sl_cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	log.Println("conf:", sl_cfg)
	return nil
}
