// zklock
package zkhelper

import (
	//"fmt"
	"strconv"
	"strings"

	"github.com/samuel/go-zookeeper/zk"
)

type LockMode int32

const (
	LockModeUnknown LockMode = 0
	LockModeRead    LockMode = 1
	LockModeWrite   LockMode = 2
	LockModeMutex   LockMode = 3
)

/****************************************
 * zk_get_sequential_min_num
 * get the sequential min node
 *
 ****************************************/
func zk_get_sequential_min_num(paths []string, mode LockMode) (int, error) {
	//var node_before string
	//var err error
	var min int = -1
	if len(paths) == 0 {
		return min, nil
	}
	for _, path := range paths {
		switch mode {
		case LockModeRead:
			bwrite := strings.Contains(path, "write")
			if !bwrite {
				continue
			}
			break
		default:
		}
		num, _ := zk_get_sequential_num(path)
		if -1 == min || num < min {
			min = num
		}

	}
	return min, nil
}

func zk_get_sequential_num(path string) (int, error) {
	var num int = 0
	log.Println("str:", path)
	lastIndex := strings.LastIndex(path, "-")
	snum := path[lastIndex+1:]
	log.Println("str:", snum)
	num, err := strconv.Atoi(snum)
	log.Printf("num:%v\n", num)
	if err != nil {
		return num, err
	}
	return num, nil
}

/****************************************
 * zk_lock
 * creat EPHEMERAL_SEQUENTIAL node
 *
 ****************************************/
func (client *ZKClient) Lock(node *ZKNode, mode LockMode) (error, <-chan *ZKNode) {
	var path string
	if err, _ := client.PathExist(node.Path); err != nil {
		return err, nil
	}
	//	var lockroot string = node.Path
	switch mode {
	case LockModeRead:
		path = node.Path + "/" + node.Name + "-"
		break
	case LockModeWrite:
		path = node.Path + "/write-" + node.Name + "-"
		break
	case LockModeMutex:
		path = node.Path + "/mutex-" + node.Name + "-"
		break
	default:
	}
	path, err := client.conn.Create(path, nil, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err, nil
	}

	for {
		children, err := client.GetChildren(node)
		if err != nil {
			log.Errorln(err)
		}
		if len(children) == 0 {
			break
		}
		minnum, _ := zk_get_sequential_min_num(children, mode)
		num, _ := zk_get_sequential_num(path)
		log.Infof("num: %d, min: %d", num, minnum)
		if num <= minnum {
			break
		}
	}
	node.Path = path
	ch := make(chan *ZKNode, 1)
	ch <- node
	return nil, ch
}

/****************************************
 * zk_unlock
 * delete node
 *
 ****************************************/
func (client *ZKClient) UnLock(node *ZKNode) error {
	err, s := client.PathExist(node.Path)
	if err != nil {
		log.Errorf("Node %s do not exist! err:%s", node.Path, err)
		return err
	}
	err = client.conn.Delete(node.Path, s.Version)
	if err != nil {
		log.Errorf("Node %s unlock failed! err:%s", node.Path, err)
		return err
	}
	return nil
}
