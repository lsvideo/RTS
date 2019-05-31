// eChatDVR
package main

import (
	"encoding/json"
	"io/ioutil"
	"path"
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

var eChatVodNode zkhelper.ZKNode

func init() {
	RegisterCommand("eChatDvr", EchatDvr)
	RegisterCommand("eChatDelVodFile", EchatDelVodFile)
	DiskManager()
	eChatVodNode.ServiceType = zkhelper.GetServiceType(config.Type)
	eChatVodNode.Name = zkhelper.SHANLI_ZK_FUNC_VOD
	eChatVodNode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(eChatVodNode.ServiceType), zkhelper.NodeTypeVod)
}

func (dvrInfo *eChatDvr) store(user *srs_eChatUser) error {

	var companynode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	companynode.SetServiceType(zkhelper.ServiceTypeRTMP)
	companynode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(companynode.ServiceType), zkhelper.NodeTypeVod) + "/" + user.User.Cid)
	companynode.SetName(user.User.Cid)
	exists, _, err := rtsclient.Exist(&companynode)
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
	usernode.SetPath(usernode.Path + "/" + user.Stream)
	usernode.SetName(user.Stream)
	exists, _, err = rtsclient.Exist(&usernode)
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

	exists, _, err = rtsclient.Exist(&typenode)
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
	log.Infoln("EchatDvr:", srsuser)

	filename := filepath.Base(srsuser.Dvr_File)
	log.Infoln("filename:", filename)
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
		log.Errorln("EchatDvr :", srsuser, " err:", err)
	}

}

func DiskManager() {
	cronInit()
	crontask.AddFunc("@hourly", DiskClean)
	//crontask.AddFunc("*/60 * * * * *", DiskClean)
}

func DiskClean() {
	log.Println("DiskClean")
	for {
		if dfpercent := DiskUsage(config.Dvr_path); dfpercent > 0.9 {
			log.Println("UsagePercent:", dfpercent)
			var task Task
			task.Task_command = "eChatDelVodFile"
			AddTask(task)
		} else {
			log.Println("UsagePercent:", dfpercent)
			break
		}

	}
}

func EchatDelVodFile(t Task) {
	var strOld string = ""
	var iOld, iNew int64 = 0, 0
	dir_list, e := ioutil.ReadDir(config.Dvr_path)
	if e != nil {
		log.Println("read dir error")
		return
	}
	for i, v := range dir_list {
		log.Println(i, "=", v.Name(), " Ext: ", path.Ext(v.Name()))
		if ext := path.Ext(v.Name()); ext != ".flv" {
			continue
		}
		if strOld == "" {
			strOld = v.Name()
			iOld = EchatGetFileCreatTime(strOld)
			continue
		}
		iNew = EchatGetFileCreatTime(v.Name())
		if iNew < iOld {
			strOld = v.Name()
			iOld = iNew
		}
	}
	log.Println("Oldest file:", strOld)
	if strOld != "" {
		EchatDeleteVodNode(strOld)
		err := RemoveFile(config.Dvr_path + "/" + strOld)
		if err != nil {
			log.Errorln("Delete file: ", config.Dvr_path, "/", strOld, " err: ", err)
		}
	}
}

func EchatGetFileCreatTime(filename string) (createTime int64) {
	indexStart := strings.Index(filename, ".")
	indexEnd := strings.LastIndex(filename, ".")
	if indexStart < indexEnd {
		strTime := filename[indexStart+1 : indexEnd]
		log.Println("EchatGetFileCreatTime:", strTime)
		createTime, _ = strconv.ParseInt(strTime, 0, 64)
	} else {
		createTime = 0
	}
	return
}

func EchatDeleteVodNode(filename string) {
	log.Infoln("EchatDeleteVodNode:", filename)
	strarr := strings.Split(filename, ".")
	uid := strarr[0]

	children, err := rtsclient.GetChildren(&eChatVodNode)
	if err != nil {
		log.Errorln(err)
	}
	if len(children) != 0 {
		for _, name := range children {
			log.Infoln("child:", name)
			var childnode zkhelper.ZKNode = eChatVodNode
			childnode.SetPath(eChatVodNode.Path + "/" + name + "/" + uid)
			log.Infoln("Chile uid node :", childnode.Path)
			bExist, _, _ := rtsclient.Exist(&childnode)
			if bExist {
				log.Infoln("uid node :", childnode.Path)
				var delnode zkhelper.ZKNode = eChatVodNode
				delnode.SetPath(childnode.Path + "/" + filename)
				log.Infoln("EchatDeleteVodNode delete:", delnode.Path)
				err := rtsclient.Delete(&delnode)
				if err != nil {
					log.Errorln("EchatDeleteVodNode err:", err)
				}
				break
			}
		}
	}
}
