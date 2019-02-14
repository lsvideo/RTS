// zknode
package zkhelper

import (
	"strings"
)

type NodeType int32

const (
	NodeTypeUnknown      NodeType = 0
	NodeTypeAutoDetected NodeType = 1
	NodeTypeLock         NodeType = 2
	NodeTypeServer       NodeType = 3
	NodeTypeUser         NodeType = 4
)

var (
	nodeTypeNames = map[NodeType]string{
		NodeTypeUnknown:      "NodeTypeUnknown",
		NodeTypeAutoDetected: "NodeTypeAutoDetected",
		NodeTypeLock:         "NodeTypeLock",
		NodeTypeServer:       "NodeTypeServer",
		NodeTypeUser:         "NodeTypeUser",
	}
)

type NodeStatus int32

const (
	NodeStatusUnknown NodeStatus = 0
	NodeStatusNew     NodeStatus = 1
	NodeStatusNormal  NodeStatus = 2
	NodeStatusWatched NodeStatus = 3
	NodeStatusDelete  NodeStatus = 4
)

var (
	nodeStatusNames = map[NodeStatus]string{
		NodeStatusUnknown: "NodeStatusUnknown",
		NodeStatusNormal:  "NodeStatusNormal",
		NodeStatusWatched: "NodeStatusWatched",
		NodeStatusDelete:  "NodeStatusDelete",
	}
)

type ServiceType int32

const (
	ServiceTypeUnknown        ServiceType = 0
	ServiceTypeRTMP           ServiceType = 1
	ServiceTypeClusterManager ServiceType = 2
)

var (
	serviceTypeNames = map[ServiceType]string{
		ServiceTypeUnknown:        "ServiceTypeUnknown",
		ServiceTypeRTMP:           "ServiceTypeRTMP",
		ServiceTypeClusterManager: "ServiceTypeClusterManager",
	}
)

var (
	SHANLI_SEPARATOR       = "/"
	SHANLI_ZK_ROOT         = "shanli"
	SHANLI_ZK_APP_RTMP     = "rtmp"
	SHANLI_ZK_APP_CM       = "clustermanager"
	SHANLI_ZK_APP_DIRS     = []string{SHANLI_ZK_APP_RTMP, SHANLI_ZK_APP_CM}
	SHANLI_ZK_FUNC_AUTO    = "autodetected"
	SHANLI_ZK_FUNC_SERVERS = "servers"
	SHANLI_ZK_FUNC_USERS   = "users"
	SHANLI_ZK_FUNC         = []string{SHANLI_ZK_FUNC_AUTO, SHANLI_ZK_FUNC_SERVERS, SHANLI_ZK_FUNC_USERS}
	SHANLI_ZK_LOCK         = "lock"
)

//service node
//one client can register multiple services
type ZKNode struct {
	Name        string      `json:"name"` // 服务名称，这里是 user
	NodeType    NodeType    `json:"nodetype"`
	ServiceType ServiceType `json:"servicetype"`
	Path        string      `json:"path"`
	Status      NodeStatus  `json:"status"`
	//IP          string   `json:"ip"`
	//Port        string   `json:"port"`
	NumChildren int      `json:"numchildren"`
	Children    []ZKNode `json:"child"`
	//stat        zk.Stat
	Date []byte `json:"date"`
}

func GeneratePath(paths ...string) (finalpath string) {
	for _, path := range paths {
		if 0 == strings.Index(path, "/") {
			finalpath += (path)
		} else {
			finalpath += (SHANLI_SEPARATOR + path)
		}
	}
	log.Println(finalpath)
	return finalpath
}

func GetServicePath(st ServiceType) (finalpath string) {
	switch st {
	case ServiceTypeRTMP:
		finalpath = GeneratePath(SHANLI_ZK_ROOT, SHANLI_ZK_APP_RTMP)
	case ServiceTypeClusterManager:
		finalpath = GeneratePath(SHANLI_ZK_ROOT, SHANLI_ZK_APP_CM)
	default:
		finalpath = ""
	}
	return finalpath
}

func GetNodePath(parent string, nt NodeType) (finalpath string) {
	finalpath = ""
	switch nt {
	case NodeTypeAutoDetected:
		finalpath = GeneratePath(parent, SHANLI_ZK_FUNC_AUTO)
	case NodeTypeServer:
		finalpath = GeneratePath(parent, SHANLI_ZK_FUNC_SERVERS)
	case NodeTypeUser:
		finalpath = GeneratePath(parent, SHANLI_ZK_FUNC_USERS)
	case NodeTypeLock:
		finalpath = GeneratePath(parent, SHANLI_ZK_LOCK)
	}
	return finalpath
}

func GetServiceType(st string) ServiceType {
	switch strings.ToLower(st) {
	case "rtmp":
		return ServiceTypeRTMP
	case "clustermanager":
		return ServiceTypeClusterManager
	default:
		return ServiceTypeUnknown
	}
}

func NewNode(st string, nt NodeType) *NodeType {
	return nil
}
