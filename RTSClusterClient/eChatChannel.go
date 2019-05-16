// eChatChannel
package main

import (
	"container/list"
	"encoding/json"
	"strconv"
	"sync"
	"zkhelper"
)

var mapeChatChannels sync.Map

type eChatChannel struct {
	Uid         string     `json:"uid"` //Channelid
	lstUsers    *list.List //用户列表
	channelnode *zkhelper.ZKNode
}

func init() {
	RegisterCommand("eChatAddChannel", EchatAddChannel)
	RegisterCommand("eChatDeleteChannel", EchatDeleteChannel)
	RegisterCommand("eChatAddUserToChannel", EchatAddUserToChannel)
	RegisterCommand("eChatDeleteUserFromChannel", EchatDeleteUserFromChannel)
}

func eChatCannelInit() {
	//EchatChannel 节点初始化  ../channels/IP:PORT
	var channelnode zkhelper.ZKNode
	channelnode.ServiceType = zkhelper.GetServiceType(config.Type)
	channelnode.Name = config.IP + ":" + strconv.Itoa(config.Port)
	channelnode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(channelnode.ServiceType), zkhelper.NodeTypeChannel) + "/" + channelnode.Name
	exists, err := rtsclient.Exist(&channelnode)
	if err != nil {
		log.Errorln(err)
	}
	if !exists {
		err := rtsclient.Create(&channelnode)
		if err != nil {
			log.Errorln(err)
		}
	}

}

func (c *eChatChannel) store(srsuser srs_eChatUser) error {
	c.Uid = srsuser.User.Uid
	_, ok := mapeChatChannels.Load(c.Uid)
	if !ok {
		var channelnode zkhelper.ZKNode
		channelnode.SetServiceType(zkhelper.ServiceTypeRTMP)
		channelnode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(channelnode.ServiceType), zkhelper.NodeTypeChannel) + "/" + srsuser.User.Url + "/" + c.Uid)
		channelnode.SetName(c.Uid)
		buf, _ := json.Marshal(srsuser)
		channelnode.SetData(buf)
		exists, err := rtsclient.Exist(&channelnode)
		if err != nil {
			return err
		}
		if !exists {
			err := rtsclient.Create(&channelnode)
			if err != nil {
				return err
			}
		}
		c.channelnode = &channelnode
		c.lstUsers = list.New()
		mapeChatChannels.Store(c.Uid, c)
	}
	return nil
}

func EchatAddChannel(t Task) {
	var echatchannel eChatChannel
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@EchatAddChannel:", srsuser)

	err := echatchannel.store(srsuser)
	if err != nil {
		log.Errorln("@@@@EchatAddChannel :", echatchannel, " err:", err)
	}

	//mapChannels.Range(sc)
	mapeChatChannels.Range(MapGoThrough)

}

func deleChatChannel(c *eChatChannel) error {
	var channel *eChatChannel
	mapeChatChannels.Range(MapGoThrough)
	value, ok := mapeChatChannels.Load(c.Uid)
	if ok {
		channel = value.(*eChatChannel)
		delUserfromeChatChannel(channel, "all")
		err := rtsclient.Delete(channel.channelnode)
		if err != nil {
			return err
		}
		mapeChatChannels.Delete(channel.Uid)
	} else {
		log.Errorln("!!!!!!!!!!!!!!")
	}

	return nil
}

func EchatDeleteChannel(t Task) {
	var channel eChatChannel
	json.Unmarshal([]byte(t.Task_data), &channel)
	log.Infoln("@@@@@@@@@@@@@@@@@@@@@@@@EchatDeleteChannel:", channel)
	err := deleChatChannel(&channel)
	if err != nil {
		log.Errorln("@@@@@@@@@@@@@@@@@@@@@@EchatDeleteChannel :", channel, " err:", err)
	}

}

func echatUserInChannel(channel *eChatChannel, uid string) bool {

	var bExist bool = false
	for e := channel.lstUsers.Front(); e != nil; e = e.Next() {
		user := e.Value.(string)
		if user == uid {
			bExist = true
		}
	}
	return bExist
}

func EchatAddUserToChannel(t Task) {
	var channel eChatChannel
	var srsuser srs_eChatUser
	var echatuser *eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@EchatAddUserToChannel:", srsuser, "to", t.User_id)
	channel.Uid = t.User_id
	echatuser = srsuser.User
	value, ok := mapeChatChannels.Load(channel.Uid)
	if ok {
		echatchannel := value.(*eChatChannel)
		log.Infoln("11111111111111:", *echatchannel)
		if !echatUserInChannel(echatchannel, echatuser.Uid) {
			echatchannel.lstUsers.PushBack(echatuser.Uid)
			node := *echatchannel.channelnode
			node.SetPath(node.Path + "/" + echatuser.Uid)
			node.SetData([]byte(t.Task_data))
			rtsclient.Create(&node)
			mapeChatChannels.Store(channel.Uid, echatchannel)
		}
		log.Infoln("22222222222222:", *echatchannel)
		mapeChatChannels.Range(MapGoThrough)
	}
}

func delUserfromeChatChannel(channel *eChatChannel, uid string) {
	var next *list.Element
	for e := channel.lstUsers.Front(); e != nil; e = next {
		user := e.Value.(string)
		next = e.Next()
		if user == uid {
			channel.lstUsers.Remove(e)
			node := *channel.channelnode
			node.SetPath(node.Path + "/" + uid)
			rtsclient.Delete(&node)
			break
		} else if "all" == uid {
			channel.lstUsers.Remove(e)
			node := *channel.channelnode
			node.SetPath(node.Path + "/" + user)
			rtsclient.Delete(&node)
			var task Task
			var echatchannel eChatChannel
			echatchannel.Uid = channel.Uid
			task.User_id = user
			task.Task_command = "video.StreamStop"
			buf, _ := json.Marshal(echatchannel)
			task.Task_data = string(buf)
			AddTask(task)
		}
	}
}

func EchatDeleteUserFromChannel(t Task) {
	var channel eChatChannel
	json.Unmarshal([]byte(t.Task_data), &channel)
	log.Infoln("@@@@EchatDeleteUserFromChannel:", channel)
	value, ok := mapeChatChannels.Load(channel.Uid)
	uid := t.User_id
	if ok {
		echatchannel := value.(*eChatChannel)
		delUserfromeChatChannel(echatchannel, uid)
		mapeChatChannels.Store(channel.Uid, echatchannel)
	}

}
