// zkhelper
package zkhelper

import (
	"fmt"
	//	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	//log "github.com/sirupsen/logrus"
	"sl_log"
)

var log = sl_log.Log

//zk Client
type ZKClient struct {
	zkServers []string // 多个集群管理节点地址
	//zkRoot    string   // 服务根节点，这里是 /shanlitech
	conn       *zk.Conn // zk 的客户端连接
	clienttype string   //客户端类型 rtmp clustermanager
	ip         string   //IP
	port       int      //Port
	timeout    int      //timeout
}

func Callback(event zk.Event) {
	//log.Infoln("*******************")
	//log.Infoln("path:", event.Path)
	//log.Infoln("type:", event.Type.String())
	//log.Infoln("state:", event.State.String())
	//log.Infoln("-------------------")
}

func (client *ZKClient) init_root() error {
	// 创建服务根节点
	if err, _ := client.PathExist(SHANLI_SEPARATOR + SHANLI_ZK_ROOT); err != nil {
		client.Close()
		return err
	}

	// 创建服务目录
	for _, sdir := range SHANLI_ZK_APP_DIRS {
		if err, _ := client.PathExist(SHANLI_SEPARATOR + SHANLI_ZK_ROOT + SHANLI_SEPARATOR + sdir); err != nil {
			client.Close()
			return err
		}
		for _, fdir := range SHANLI_ZK_FUNC {
			if err, _ := client.PathExist(SHANLI_SEPARATOR + SHANLI_ZK_ROOT + SHANLI_SEPARATOR + sdir + SHANLI_SEPARATOR + fdir); err != nil {
				client.Close()
				return err
			}
		}
	}
	return nil
}

func (client *ZKClient) init_service_node(node *ZKNode) error {

	switch node.ServiceType {
	case ServiceTypeRTMP:
		// 创建服务根节点
		if err, _ := client.PathExist(GetServicePath(ServiceTypeRTMP)); err != nil {
			client.Close()
			return err
		}

	}

	if err, _ := client.PathExist(node.Path); err != nil {
		return err
	}
	return nil
}

func (client *ZKClient) NewClient(zkServers []string, clienttype string, ip string, port int, timeout int) (bool, error) {
	client.zkServers = zkServers
	client.clienttype = clienttype
	client.ip = ip
	client.port = port
	client.timeout = timeout

	optioncallback := zk.WithEventCallback(Callback)
	optionlog := zk.WithLogger(log)
	// 连接服务器
	conn, _, err := zk.Connect(zkServers, time.Duration(timeout)*time.Second, optioncallback, optionlog)
	if err != nil {
		return false, err
	}
	client.conn = conn

	err = client.init_root()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (client *ZKClient) Register(node *ZKNode) error {

	if err := client.init_service_node(node); err != nil {
		return err
	}
	path := node.Path + "/" + node.Name
	data := []byte(node.Data)

	path, err := client.conn.CreateProtectedEphemeralSequential(path, data, zk.WorldACL(zk.PermAll))
	node.Path = path
	fmt.Println(node)
	if err != nil {
		return err
	}
	return nil
}

func (client *ZKClient) Create(node *ZKNode) error {
	path := node.Path
	data := []byte(node.Data)

	path, err := client.conn.Create(path, data, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}

	node.Path = path
	fmt.Println(node)
	return nil
}

func (client *ZKClient) CreateSequence(node *ZKNode) error {
	path := node.Path
	data := []byte(node.Data)

	path, err := client.conn.Create(path, data, zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}

	node.Path = path
	fmt.Println(node)
	return nil
}

func (client *ZKClient) PathExist(sPath string) (error, *zk.Stat) {
	exists, s, err := client.conn.Exists(sPath)
	if err != nil {
		return err, s
	}

	if !exists {
		_, err := client.conn.Create(sPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err, s
		}
	}
	return nil, s
}

func ZkStateString(s *zk.Stat) string {
	return fmt.Sprintf("Czxid:%d, Mzxid: %d, Ctime: %d, Mtime: %d, Version: %d, Cversion: %d, Aversion: %d, EphemeralOwner: %d, DataLength: %d, NumChildren: %d, Pzxid: %d",
		s.Czxid, s.Mzxid, s.Ctime, s.Mtime, s.Version, s.Cversion, s.Aversion, s.EphemeralOwner, s.DataLength, s.NumChildren, s.Pzxid)
}

func (client *ZKClient) WatchNode(node *ZKNode, status NodeStatus) (<-chan *ZKNode, error) {
	sPath := node.Path
	for node.Status != status {
		exists, _, node_ch, err := client.conn.ExistsW(sPath)
		//_, _, _, err := client.conn.ExistsW(sPath)
		if err != nil {
			return nil, err
		}
		if exists {
			//fmt.Printf("watch Path result state[%s]\n", ZkStateString(s))

			select {
			case ch_event := <-node_ch:
				{
					//log.Infoln("path:", ch_event.Path, " type:", ch_event.Type.String())
					//fmt.Println("state:", ch_event.State.String())
					if ch_event.Type == zk.EventNodeCreated {
						//log.Infoln("has new node[%s] create\n", ch_event.Path)
					} else if ch_event.Type == zk.EventNodeDeleted {
						log.Infoln("has node[%s] detete\n", ch_event.Path)
						node.Status = NodeStatusDelete
						ch := make(chan *ZKNode, 1)
						ch <- node
						return ch, nil
					} else if ch_event.Type == zk.EventNodeDataChanged {
						//log.Infoln("has node[%s] data changed\n", ch_event.Path)
					} else if ch_event.Type == zk.EventNodeChildrenChanged {
						//log.Infoln("has children[%s] data changed", ch_event.Path)
					}
				}
			}
		}
	}
	return nil, nil
}

func (client *ZKClient) GetChildren(node *ZKNode) ([]string, error) {
	sPath := node.Path
	children, _, err := client.conn.Children(sPath)
	//_, _, _, err := client.conn.ChildrenW(sPath)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("watch childs result state[%s] childs:\n", ZkStateString(s))
	//fmt.Println(children)
	return children, nil
}

func (client *ZKClient) WatchChild(node *ZKNode) error {
	sPath := node.Path
	_, _, node_ch, err := client.conn.ChildrenW(sPath)
	if err != nil {
		return err
	}

	//fmt.Printf("watch childs result state[%s] childs:\n", ZkStateString(s))
	//fmt.Println(children)

	select {
	case ch_event := <-node_ch:
		{
			fmt.Println("path:", ch_event.Path)
			fmt.Println("type:", ch_event.Type.String())
			fmt.Println("state:", ch_event.State.String())
			if ch_event.Type == zk.EventNodeCreated {
				fmt.Printf("has node[%s] detete\n", ch_event.Path)
			} else if ch_event.Type == zk.EventNodeDeleted {
				fmt.Printf("has new node[%s] create\n", ch_event.Path)
			} else if ch_event.Type == zk.EventNodeDataChanged {
				fmt.Printf("has node[%s] data changed", ch_event.Path)
			} else if ch_event.Type == zk.EventNodeChildrenChanged {
				fmt.Printf("has children[%s] data changed", ch_event.Path)
			}
		}
	}
	time.Sleep(time.Millisecond * 10)

	return nil
}

func (client *ZKClient) Close() {
	client.conn.Close()
}

func (client *ZKClient) Get(node *ZKNode) (string, error) {
	err, _ := client.PathExist(node.Path)
	if err != nil {
		return "", err
	}
	buf, _, err := client.conn.Get(node.Path)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (client *ZKClient) Set(node *ZKNode) {
	err, s := client.PathExist(node.Path)
	_, err = client.conn.Set(node.Path, node.Data, s.Version)
	if err != nil {
		log.Errorln(err)
	}
}

func (client *ZKClient) Delete(node *ZKNode) error {
	path := node.Path
	_, s, err := client.Exist(node)

	err = client.conn.Delete(path, s.Version)
	if err != nil {
		return err
	}

	return nil
}

func (client *ZKClient) Exist(node *ZKNode) (bool, *zk.Stat, error) {
	exists, s, err := client.conn.Exists(node.Path)

	return exists, s, err
}
