// eChatDVR
package main

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"zkhelper"
)

type eChatDvr struct {
	Url      string    `json:"url"`
	Name     string    `json:"name"`
	Start    int64     `json:"start"`
	Duration int64     `json:"duration"`
	Size     int64     `json:"size"`
	Status   DvrStatus `json:"status"`
}

type DvrStatus int32

const (
	DvrStatusVod     DvrStatus = 0
	DvrStatusdelete  DvrStatus = -1
	DvrStatusdeleted DvrStatus = -2
)

var eChatVodNode zkhelper.ZKNode
var mapeChatDvrInfo sync.Map

func init() {
	RegisterCommand("eChatDvr", EchatDvr)
	RegisterCommand("eChatDelOldestFile", EchatDelOldestFile)
	RegisterCommand("eChatDelInvalidFile", EchatDelInvalidFile)
	RegisterCommand("eChatAddDvrInfo", EchatAddDvrInfo)
	DiskManager()
	eChatVodNode.ServiceType = zkhelper.GetServiceType(config.Type)
	eChatVodNode.Name = zkhelper.SHANLI_ZK_FUNC_VOD
	eChatVodNode.Path = zkhelper.GetNodePath(zkhelper.GetServicePath(eChatVodNode.ServiceType), zkhelper.NodeTypeVod)
}

func (dvrInfo *eChatDvr) store(user *srs_eChatUser) error {
	dvrInfo.addDvrtoDB(user)

	var companynode zkhelper.ZKNode //注册公司节点
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

	var usernode zkhelper.ZKNode //注册流节点
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
	var typenode zkhelper.ZKNode //视频记录节点
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

func (dvrInfo *eChatDvr) addDvrtoDB(user *srs_eChatUser) {
	insertDvrItem(user, dvrInfo)
}

func EchatDvr(t Task) {
	var srsuser srs_eChatUser
	var dvr eChatDvr
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("EchatDvr:", srsuser)

	value, ok := mapeChatDvrInfo.Load(srsuser.Stream)
	if ok {
		user := value.(*srs_eChatUser)
		srsuser.User = user.User
		log.Infoln("EchatDvr get stream user Info:", srsuser, srsuser.User)
	}

	filename := filepath.Base(srsuser.Dvr_File)
	log.Infoln("filename:", filename)
	dvr.Name = filename
	strarr := strings.Split(filename, ".")
	dvr.Start, _ = strconv.ParseInt(strarr[1], 0, 64)
	//dvr.Url = config.IP + ":" + "8090"
	//sl.yaml 配置文件中读取
	dvr.Url = config.IP + ":" + strconv.Itoa(config.Dvr_port)
	dvr.Size = GetFileSize(srsuser.Dvr_File)
	//需安装mediainfo程序
	str, err := Exec_shell("mediainfo --Inform='General;%Duration%' " + srsuser.Dvr_File)
	dvr.Duration, _ = strconv.ParseInt(str, 0, 64)
	dvr.Status = DvrStatusVod
	//move file to dvrPath
	log.Infoln("DVR :", dvr)

	err = MoveFile(srsuser.Dvr_File, config.Dvr_path+"/"+filename)
	if err != nil {
		log.Errorln("MoveFile :", srsuser, " err:", err)
	}

	err = dvr.store(&srsuser)
	if err != nil {
		log.Errorln("EchatDvr :", srsuser, " err:", err)
	}

	mapeChatDvrInfo.Delete(srsuser.Stream)

}

func DiskManager() {
	cronInit()
	//每隔一小时检测磁盘使用
	crontask.AddFunc("@hourly", DiskClean)
	//crontask.AddFunc("*/10 * * * * *", DiskClean)
}

func DiskClean() {
	log.Infoln("DiskClean")
	//delete the file whose status is delete in tb_video_file
	var task Task
	task.Task_command = "eChatDelInvalidFile"
	EchatDelInvalidFile(task)

	//when the disk usage > 90% , delete the oldest file.
	for {
		if dfpercent := DiskUsage(config.Dvr_path); dfpercent > 0.9 {
			log.Infoln("UsagePercent:", dfpercent)
			var task Task
			task.Task_command = "eChatDelOldestFile"
			//AddTask(task)
			EchatDelOldestFile(task)
		} else {
			log.Infoln("UsagePercent:", dfpercent)
			break
		}
	}
}

func EchatDelInvalidFile(t Task) {
	server := config.IP + ":" + strconv.Itoa(config.Dvr_port)
	mapDvr := make(map[int]string)
	queryDvrIdNamebyStatus(DvrStatusdelete, server, mapDvr)
	for key, value := range mapDvr {
		filename := value
		id := key
		err := RemoveFile(config.Dvr_path + "/" + filename)
		if err != nil {
			log.Errorln("Delete file: ", config.Dvr_path, "/", filename, " err: ", err)
		}
		updateDvrStatus(id, DvrStatusdeleted)
		EchatDeleteVodNode(filename)
	}
}

func EchatUpdateDvrStatusByName(filename string, status DvrStatus) {
	server := config.IP + ":" + strconv.Itoa(config.Dvr_port)
	mapDvr := make(map[int]string)
	queryDvrIdbyName(filename, server, mapDvr)
	for key, _ := range mapDvr {
		id := key
		updateDvrStatus(id, status)
	}
}

func EchatDelOldestFile(t Task) {
	var strOld string = ""
	var iOld, iNew int64 = 0, 0
	dir_list, e := ioutil.ReadDir(config.Dvr_path)
	if e != nil {
		log.Errorln("read dir error!")
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
		err := RemoveFile(config.Dvr_path + "/" + strOld)
		if err != nil {
			log.Errorln("Delete file: ", config.Dvr_path, "/", strOld, " err: ", err)
		} else {
			EchatUpdateDvrStatusByName(strOld, DvrStatusdeleted)
			EchatDeleteVodNode(strOld)
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
			log.Infoln("Child uid node :", childnode.Path)
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

func EchatAddDvrInfo(t Task) {
	var srsuser srs_eChatUser
	json.Unmarshal([]byte(t.Task_data), &srsuser)
	log.Infoln("EchatAddDvrInfo:", srsuser)
	mapeChatDvrInfo.Store(srsuser.Stream, &srsuser)
}
