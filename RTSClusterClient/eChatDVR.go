// eChatDVR
package main

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"zkhelper"
)

type eChatDvr struct {
	Url      string `json:"url"`
	Start    int64  `json:"start"`
	Duration int64  `json:"duration"`
	Size     int64  `json:"size"`
}

func init() {
	RegisterCommand("eChatDvr", EchatDvr)
}

func (dvrInfo *eChatDvr) store(user *srs_eChatUser) error {

	var companynode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	companynode.SetServiceType(zkhelper.ServiceTypeRTMP)
	companynode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(companynode.ServiceType), zkhelper.NodeTypeVod) + "/" + user.User.Cid)
	companynode.SetName(user.User.Cid)
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
	usernode.SetPath(usernode.Path + "/" + user.User.Uid)
	usernode.SetName(user.User.Uid)
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

	filename := filepath.Base(user.Dvr_File)
	var typenode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	typenode = usernode
	typenode.SetPath(typenode.Path + "/" + filename)
	typenode.SetName(filename)
	buf, err := json.Marshal(dvrInfo)
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

	return nil
}

func EchatDvr(t Task) {
	var srsuser srs_eChatUser
	var dvr eChatDvr
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("@@@@EchatDvr:", srsuser)

	filename := filepath.Base(srsuser.Dvr_File)
	log.Infoln("@@@@filename:", filename)
	strarr := strings.Split(filename, ".")
	dvr.Start, _ = strconv.ParseInt(strarr[1], 0, 64)
	dvr.Url = config.IP + ":" + "8090"
	dvr.Size = GetFileSize(srsuser.Dvr_File)
	str, err := Exec_shell("mediainfo --Inform='General;%Duration%' " + srsuser.Dvr_File)
	dvr.Duration, _ = strconv.ParseInt(str, 0, 64)
	//move file to dvrPath
	log.Errorln("DVR :", dvr)

	err = MoveFile(srsuser.Dvr_File, config.Dvr_path+"/"+filename)
	if err != nil {
		log.Errorln("MoveFile :", srsuser, " err:", err)
	}

	err = dvr.store(&srsuser)
	if err != nil {
		log.Errorln("@@@@EchatAddUser :", srsuser, " err:", err)
	}

}
