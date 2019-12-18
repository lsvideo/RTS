// config
package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type SL_config struct {
	IP                string
	Port              int
	Type              string
	Device            string
	ZK_servers        []string
	Penetrate_server  string
	Video_server      string
	Dvr_port          int
	Dvr_path          string
	config_file       string
	Srs_callback_port int
	Srs_api_port      int
	DB_server         string
	DB_port           int
	DB_user           string
	DB_pwd            string
	DB_name           string
}

var (
	SRS_CALLBACK_DEFAULT_PORT = 10002
	SRS_API_DEFAULT_PORT      = 1985
	SRS_DVR_DEFAULT_PORT      = 8090
)

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

// init default config
func (sl_cfg *SL_config) init_config() {
	sl_cfg.Dvr_port = SRS_DVR_DEFAULT_PORT
	sl_cfg.Srs_api_port = SRS_API_DEFAULT_PORT
	sl_cfg.Srs_callback_port = SRS_CALLBACK_DEFAULT_PORT
}

func (sl_cfg *SL_config) Parse_config(cfg string) error {
	sl_cfg.config_file = cfg
	sl_cfg.init_config()
	yamlFile, err := ioutil.ReadFile(sl_cfg.config_file)
	//log.Println("yamlFile:", yamlFile)
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
