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

type srs_eChatUser struct {
	Client_id int        `json:"client_id"`
	Stream    string     `json:"stream"`
	User      *eChatUser `json:"echat_user"`
}

func init() {
	RegisterCommand("eChatAddUser", EchatAddUser)
	RegisterCommand("eChatDeleteUser", EchatDeleteUser)
}

func (u *eChatUser) getUidMark(mark string) string {
	return u.Uid + "-" + u.Option + "-" + mark
}

func (u *srs_eChatUser) store() error {
	var companynode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	companynode.SetServiceType(zkhelper.ServiceTypeRTMP)
	companynode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(companynode.ServiceType), zkhelper.NodeTypeUser) + "/" + u.User.Cid)
	companynode.SetName(u.User.Cid)
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
	usernode.SetPath(usernode.Path + "/" + u.User.Uid)
	usernode.SetName(u.User.Uid)
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
	typenode.SetPath(typenode.Path + "/" + u.User.Action + "_" + u.Stream)
	typenode.SetName(u.User.Action)
	buf, err := json.Marshal(u.User)
	if err != nil {
		return err
	}
	typenode.Data = buf
	//err = rtsclient.CreateSequence(&typenode)

	exists, err = rtsclient.Exist(&typenode)
	if err != nil {
		return err
	}
	if !exists {
		err := rtsclient.Create(&typenode)
		if err != nil {
			return err
		}
	} else {
		rtsclient.Set(&typenode)
	}

	u.User.usernode = &typenode
	mapeChatUser.Store(u.User.getUidMark(u.Stream), u)

	return nil
}

func EchatAddUser(t Task) {
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@EchatAddUser:", srsuser)
	err := srsuser.store()
	if err != nil {
		log.Errorln("@@@@EchatAddUser :", srsuser, " err:", err)
	}

}

func delsrsuser(u *srs_eChatUser) error {
	var user *srs_eChatUser
	mapeChatUser.Range(MapGoThrough)
	value, ok := mapeChatUser.Load(u.User.getUidMark(u.Stream))
	if ok {
		user = value.(*srs_eChatUser)
		err := rtsclient.Delete(user.User.usernode)
		if err != nil {
			return err
		}
		mapeChatUser.Delete(user.User.getUidMark(u.Stream))
	} else {
		log.Errorln("!!!!!!!!!!!!!!")
	}

	return nil
}

func EchatDeleteUser(t Task) {
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("eChatUser:", srsuser)
	err := delsrsuser(&srsuser)
	if err != nil {
		log.Errorln("EchatDeleteUser :", srsuser, " err:", err)
	}

}
