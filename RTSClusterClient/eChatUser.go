// eChatAccess
package main

import (
	"encoding/json"
	"sync"
	"zkhelper"
)

var mapeChatUser sync.Map

type eChatUser struct {
	Uid    string `json:"uid"`  //用户id
	Cid    string `json:"cid"`  //公司id
	Action string `json:"type"` //行为类型
	Option string `json:"opt"`  //视频操作类型
	Url    string `json:"url"`  //用户使用的URL
	//	DstUid   string `json:"dstuid"` //目标uid
	usernode *zkhelper.ZKNode
}

func init() {
	RegisterCommand("eChatAddUser", EchatAddUser)
	RegisterCommand("eChatDeleteUser", EchatDeleteUser)
}

func (u *eChatUser) getUidMark() string {
	return u.Uid + "=" + u.Option
}

func (u *eChatUser) store() error {
	var companynode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	companynode.SetServiceType(zkhelper.ServiceTypeRTMP)
	companynode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(companynode.ServiceType), zkhelper.NodeTypeUser) + "/" + u.Cid)
	companynode.SetName(u.Cid)
	exists, err := rtsclient.Exist(&companynode)
	if err != nil {
		return err
	}
	if !exists {
		err := rtsclient.Create(&companynode)
		if err != nil {
			return err
		}
	}

	var usernode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	usernode = companynode
	usernode.SetPath(usernode.Path + "/" + u.Uid)
	usernode.SetName(u.Uid)
	exists, err = rtsclient.Exist(&usernode)
	if err != nil {
		return err
	}
	if !exists {
		err := rtsclient.Create(&usernode)
		if err != nil {
			return err
		}
	}

	var typenode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	typenode = usernode
	typenode.SetPath(typenode.Path + "/" + u.Action + "_")
	typenode.SetName(u.Action)
	buf, err := json.Marshal(u)
	if err != nil {
		return err
	}
	typenode.Data = buf
	err = rtsclient.CreateSequence(&typenode)
	if err != nil {
		return err
	}

	u.usernode = &typenode
	mapeChatUser.Store(u.getUidMark(), u)

	return nil
}

func EchatAddUser(t Task) {
	var echatuser eChatUser
	json.Unmarshal([]byte(t.Task_data), &echatuser)
	log.Infoln("@@@@EchatAddUser:", echatuser)
	err := echatuser.store()
	if err != nil {
		log.Errorln("@@@@EchatAddUser :", echatuser, " err:", err)
	}

}

func deluser(u *eChatUser) error {
	var user *eChatUser
	mapeChatUser.Range(MapGoThrough)
	value, ok := mapeChatUser.Load(u.getUidMark())
	if ok {
		user = value.(*eChatUser)
		log.Errorln("!11111", user)
		err := rtsclient.Delete(user.usernode)
		if err != nil {
			return err
		}
		log.Errorln("!22222", user)
		mapeChatUser.Delete(u.getUidMark())
	} else {
		log.Errorln("!!!!!!!!!!!!!!")
	}

	return nil
}

func EchatDeleteUser(t Task) {
	var echatuser eChatUser
	json.Unmarshal([]byte(t.Task_data), &echatuser)
	log.Infoln("@@@@eChatUser:", echatuser)
	err := deluser(&echatuser)
	if err != nil {
		log.Errorln("@@@@EchatAddUser :", echatuser, " err:", err)
	}

}
