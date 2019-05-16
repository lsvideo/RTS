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
var eChatChannelNode zkhelper.ZKNode

type eChatChannel struct {
	Uid         string     `json:"uid"` //Channelid
	lstUsers    *list.List //用户列表
	channelnode *zkhelper.ZKNode
}

func (u *eChatChannel) getUidMark(user string, opt string) string {
	return user + "-" + opt + "-" + u.Uid
}

func init() {
	RegisterCommand("eChatAddChannel", EchatAddChannel)
	RegisterCommand("eChatDeleteChannel", EchatDeleteChannel)
	RegisterCommand("eChatAddUserToChannel", EchatAddUserToChannel)
	RegisterCommand("eChatDeleteUserFromChannel", EchatDeleteUserFromChannel)
}

func eChatCannelInit() {
	eChatChannelNodeInit()
	eChatDataInit()
	eChatDataCheck()
}

func eChatChannelNodeInit() {
	//EchatChannel 节点初始化  ../channels/IP:PORT
	eChatChannelNode.ServiceType = zkhelper.GetServiceType(config.Type)
	eChatChannelNode.Name = config.IP + ":" + strconv.Itoa(config.Port)
	eChatChannelNode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(eChatChannelNode.ServiceType), zkhelper.NodeTypeChannel) + "/" + eChatChannelNode.Name
	exists, err := rtsclient.Exist(&eChatChannelNode)
	if err != nil {
		log.Errorln(err)
	}
	if !exists {
		err := rtsclient.Create(&eChatChannelNode)
		if err != nil {
			log.Errorln(err)
		}
	}
}

func getSrsUsers(node *zkhelper.ZKNode) {
	children, err := rtsclient.GetChildren(node)
	if err != nil {
		log.Errorln(err)
	}
	log.Infoln("node:", node.Path)
	if len(children) != 0 {
		for _, name := range children {
			log.Infoln("child:", name)
			var childnode zkhelper.ZKNode = *node
			childnode.SetPath(node.Path + "/" + name)
			str, err := rtsclient.Get(&childnode)
			if err != nil {
				log.Infoln("Get node err:", err)
			} else {
				log.Infoln("Get node :", str)
				var task Task
				task.Task_data = str
				EchatAddUser(task)
				if node.Path == eChatChannelNode.Path {
					EchatAddChannel(task)
				} else {
					EchatAddUserToChannel(task)
				}
			}
			getSrsUsers(&childnode)
		}
	}
}

func eChatDataInit() {
	//read data under eChatChannelNode
	getSrsUsers(&eChatChannelNode)
	log.Infoln("All Channel :")
	mapeChatChannels.Range(MapGoThrough)
	log.Infoln("All User :")
	mapeChatUser.Range(MapGoThrough)

}

func mapeChatChannelCheck(k, v interface{}) bool {
	uid := k.(string)
	channel := v.(*eChatChannel)
	log.Infoln("channel :", uid, " - ", *channel)
	value, _ := mapeChatUser.Load(channel.getUidMark(uid, "1"))
	srsuser := value.(*srs_eChatUser)
	client_info, err := get_client_info("127.0.0.1:1985", strconv.Itoa(srsuser.Client_id))
	if client_info == nil || client_info.Code != 0 || err != nil {
		log.Errorln("client_info :", client_info, " - ", err)
		var task Task
		buf, _ := json.Marshal(srsuser)
		task.Task_data = string(buf)
		EchatDeleteUser(task)
		EchatDeleteChannel(task)
	} else {
		log.Errorln("client_info :", client_info, " - ", err)
		var next *list.Element
		for e := channel.lstUsers.Front(); e != nil; e = next {
			useruid := e.Value.(string)
			value, _ := mapeChatUser.Load(channel.getUidMark(useruid, "2"))
			channeluser := value.(*srs_eChatUser)
			user_client_info, err := get_client_info("127.0.0.1:1985", strconv.Itoa(channeluser.Client_id))
			log.Errorln("user_client_info :", user_client_info, " - ", err)
			log.Errorln("channeluser :", channeluser)
			if user_client_info == nil || user_client_info.Code != 0 || err != nil {
				log.Errorln("44444444user_client_info :", user_client_info, " - ", err)
				var task Task
				buf, _ := json.Marshal(channeluser)
				task.Task_data = string(buf)
				EchatDeleteUser(task)
				EchatDeleteUserFromChannel(task)
			}
		}
	}
	return true
}

func eChatDataCheck() {
	//Check data under eChatChannelNode
	mapeChatChannels.Range(mapeChatChannelCheck)
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
	//	mapeChatChannels.Range(MapGoThrough)
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
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@@@@@@@@@@@@@@@@@@@@@EchatDeleteChannel:", srsuser)
	channel.Uid = srsuser.User.Uid
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
	channel.Uid = srsuser.Stream
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
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@EchatDeleteUserFromChannel:", srsuser)
	channel.Uid = srsuser.Stream
	value, ok := mapeChatChannels.Load(channel.Uid)
	uid := srsuser.User.Uid
	if ok {
		echatchannel := value.(*eChatChannel)
		delUserfromeChatChannel(echatchannel, uid)
		mapeChatChannels.Store(channel.Uid, echatchannel)
	}

}
