// srs
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var Conns uint32

type summaries struct {
	Code int          `json:"code"` // 名称
	Date summary_info `json:"data"`
}

type summary_info struct {
	Status  bool         `json:"ok"` //
	Systime int64        `json:"now_ms"`
	Program program_info `json:"self"`   //
	System  system_info  `json:"system"` //
}

type program_info struct {
	Version     string  `json:"version"`
	Argv        string  `json:"argv"`
	Cwd         string  `json:"cwd"`
	Pid         int     `json:"pid"`
	Ppid        int     `json:"ppid"`
	Mem_kbyte   int     `json:"mem_kbyte"`
	Mem_percent float32 `json:"mem_percent"`
	Cpu_percent float32 `json:"cpu_percent"`
	Srs_uptime  float32 `json:"srs_uptime"`
}

type system_info struct {
	Conn_srs int `json:"conn_srs"`
}

type SrsRTMP struct {
	Action     string `json:"action"` // 名称
	Client_id  int    `json:"client_id"`
	IP         string `json:"ip"`
	Vhost      string `json:"vhost"`
	App        string `json:"app"`
	Stream     string `json:"stream"`
	TcUrl      string `json:"tcUrl"`
	PageUrl    string `json:"pageUrl"`
	Param      string `json:"param"`
	Dvr_Path   string `json:"cwd"`
	Dvr_File   string `json:"file"`
	Send_bytes int    `json:"send_bytes"`
	Recv_bytes int    `json:"recv_bytes"`
}

type srs_client struct {
	Id      int    `json:"id"`
	Vhost   int    `json:"vhost"`
	Stream  int    `json:"stream"`
	IP      string `json:"ip"`
	PageUrl string `json:"pageUrl"`
	SwfUrl  string `json:"swfUrl"`
	TcUrl   string `json:"tcUrl"`
	Url     string `json:"url"`
	Action  string `json:"type"`
	Publish bool   `json:"publish"`
	Alive   uint64 `json:"alive"`
}

type srs_client_info struct {
	Code   int        `json:"code"` // 名称
	Server int        `json:"server"`
	Client srs_client `json:"client"`
}

var (
	SRS_VHOST_VOD = "vod"
)

func get_summaries(url string) (sum *summaries, err error) {
	resp, err := http.Get("http://" + url + "/api/v1/summaries")
	if err != nil {
		//log.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		//log.Println(err)
		return nil, err
	}

	//log.Println(string(body))

	var data summaries
	sum = &data
	err = json.Unmarshal(body, sum) //解析json格式数据
	if err != nil {
		//log.Printf("Failed unmarshalling json: %s\n", err)
		return nil, err
	}
	return sum, nil
}

func get_client_info(url string, client_id string) (client_info *srs_client_info, err error) {
	resp, err := http.Get("http://" + url + "/api/v1/clients/" + client_id)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, err
	}

	//log.Println(string(body))

	var client srs_client_info
	client_info = &client
	err = json.Unmarshal(body, client_info) //解析json格式数据
	if err != nil {
		log.Printf("Failed unmarshalling json: %s\n", err)
		return nil, err
	}
	return client_info, nil
}

func srs_connect(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.TcUrl)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		log.Infoln(m)
		log.Infoln("token:", m.Get("token"))
		log.Infoln("uid:", m.Get("uid"))
		log.Infoln("IP:", srsData.IP)

		var token Token
		uid := m.Get("uid")
		strtoken := m.Get("token")
		strip := srsData.IP

		//token 验证
		//token.GetToken(uid)
		token.GetToken("1001")
		err := token.Verify(strtoken, strip)
		if err != nil {
			log.Errorln("Token verify uid:", uid, "ip: ", strip, " err:", err)
			res = false
		}
	}
	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}

}

func srs_close(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	w.WriteHeader(200)
	w.Write([]byte("0"))
}

func srs_publish(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)

	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.TcUrl)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		log.Infoln(m)
		log.Infoln("uid:", m.Get("uid"))
		log.Infoln("cid:", m.Get("cid"))
		log.Infoln("type:", m.Get("type"))
		log.Infoln("opt:", m.Get("opt"))

		var echatuser eChatUser
		echatuser.Uid = m.Get("uid")
		echatuser.Cid = m.Get("cid")
		echatuser.Url = u.Host
		echatuser.Action = m.Get("type")
		echatuser.Option = m.Get("opt")

		var srsuser srs_eChatUser
		srsuser.Client_id = srsData.Client_id
		srsuser.Stream = srsData.Stream
		srsuser.User = &echatuser

		var task Task
		task.Task_command = "eChatAddUser"
		buf, _ := json.Marshal(srsuser)
		task.Task_data = string(buf)
		AddTask(task)

		//var echatchannel eChatChannel
		//echatchannel.Uid = m.Get("uid")
		task.Task_command = "eChatAddChannel"
		//buf, _ = json.Marshal(echatchannel)
		//task.Task_data = string(buf)
		AddTask(task)

		if srsData.Vhost == SRS_VHOST_VOD {

		}
	}

	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}
}

func srs_unpublish(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.Param)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		log.Infoln(m)
		log.Infoln("uid:", m.Get("uid"))
		log.Infoln("cid:", m.Get("cid"))
		log.Infoln("type:", m.Get("type"))
		log.Infoln("opt:", m.Get("opt"))

		var echatuser eChatUser
		echatuser.Uid = m.Get("uid")
		echatuser.Cid = m.Get("cid")
		echatuser.Url = u.Host
		echatuser.Action = m.Get("type")
		echatuser.Option = m.Get("opt")

		var srsuser srs_eChatUser
		srsuser.Client_id = srsData.Client_id
		srsuser.Stream = srsData.Stream
		srsuser.User = &echatuser

		var task Task
		task.Task_command = "eChatDeleteUser"
		buf, _ := json.Marshal(srsuser)
		task.Task_data = string(buf)
		AddTask(task)

		task.Task_command = "eChatDeleteChannel"
		AddTask(task)
	}

	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}
}

func srs_play(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.Param)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		//channel户存在为前提
		_, ok := mapeChatChannels.Load(srsData.Stream)
		if ok {
			log.Infoln(m)
			log.Infoln("uid:", m.Get("uid"))
			log.Infoln("cid:", m.Get("cid"))
			log.Infoln("type:", m.Get("type"))
			log.Infoln("opt:", m.Get("opt"))

			var echatuser eChatUser
			echatuser.Uid = m.Get("uid")
			echatuser.Cid = m.Get("cid")
			echatuser.Url = config.IP + ":" + strconv.Itoa(config.Port) //SRSManger和SRS一一对应 config中的IP:PORT 即为SRS地址
			echatuser.Action = m.Get("type")
			echatuser.Option = m.Get("opt")
			//AddTask()

			//流已推送

			var srsuser srs_eChatUser
			srsuser.Client_id = srsData.Client_id
			srsuser.Stream = srsData.Stream
			srsuser.User = &echatuser

			var task Task
			//拉流不记录
			task.Task_command = "eChatAddUser"
			buf, _ := json.Marshal(srsuser)
			task.Task_data = string(buf)
			AddTask(task)

			//var echatchannel eChatChannel
			//echatchannel.Uid = srsData.Stream
			//task.User_id = srsData.Stream
			task.Task_command = "eChatAddUserToChannel"
			//buf, _ = json.Marshal(echatchannel)
			//task.Task_data = string(buf)
			AddTask(task)
		} else {
			res = false
		}
	}

	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}
}

func srs_stop(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.Param)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		//用户存在为前提
		_, ok := mapeChatUser.Load(m.Get("uid") + "-" + m.Get("opt") + "-" + srsData.Stream)
		if ok {
			log.Infoln(m)
			log.Infoln("uid:", m.Get("uid"))
			log.Infoln("cid:", m.Get("cid"))
			log.Infoln("type:", m.Get("type"))
			log.Infoln("opt:", m.Get("opt"))

			var echatuser eChatUser
			echatuser.Uid = m.Get("uid")
			echatuser.Cid = m.Get("cid")
			echatuser.Url = u.Host
			echatuser.Action = m.Get("type")
			echatuser.Option = m.Get("opt")

			var srsuser srs_eChatUser
			srsuser.Client_id = srsData.Client_id
			srsuser.Stream = srsData.Stream
			srsuser.User = &echatuser

			var task Task
			task.Task_command = "eChatDeleteUser"
			buf, _ := json.Marshal(srsuser)
			task.Task_data = string(buf)
			AddTask(task)

			task.Task_command = "eChatDeleteUserFromChannel"
			AddTask(task)
		} else {
			res = false
		}
	}

	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}
}

func srs_dvr(w http.ResponseWriter, r *http.Request) {
	var res bool = true
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var srsData SrsRTMP
	if err := json.Unmarshal(body, &srsData); err == nil {
		log.Println("!!!!!", srsData)
	} else {
		log.Println(err)
	}
	//解析参数
	u, err := url.Parse(srsData.Param)
	if err != nil {
		res = false
	}
	log.Infoln(u.RawQuery)
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Errorln("Parse rtmp URL err:", err)
	} else {
		//用户存在为前提

		log.Infoln(m)
		log.Infoln("uid:", m.Get("uid"))
		log.Infoln("cid:", m.Get("cid"))
		log.Infoln("type:", m.Get("type"))
		log.Infoln("opt:", m.Get("opt"))

		var echatuser eChatUser
		echatuser.Uid = m.Get("uid")
		echatuser.Cid = m.Get("cid")
		echatuser.Url = config.IP + ":" + strconv.Itoa(config.Port) //SRSManger和SRS一一对应 config中的IP:PORT 即为SRS地址
		echatuser.Action = m.Get("type")
		echatuser.Option = m.Get("opt")

		var srsuser srs_eChatUser
		srsuser.Client_id = srsData.Client_id
		srsuser.Stream = srsData.Stream
		srsuser.User = &echatuser
		srsuser.Dvr_File = srsData.Dvr_Path + srsData.Dvr_File[strings.Index(srsData.Dvr_File, "/"):]
		log.Infoln("IIIIIIIIIIIIIIIIIIIIIIIIsrsuser:", srsuser)
		var task Task
		task.Task_command = "eChatDvr"
		buf, _ := json.Marshal(srsuser)
		task.Task_data = string(buf)
		AddTask(task)
	}

	res = true
	w.WriteHeader(200)
	if res {
		w.Write([]byte("0"))
	} else {
		w.Write([]byte("1"))
	}
}

func srsmanager() {
	defer PanicRecover()()
	http.HandleFunc("/srs_connect", srs_connect)
	http.HandleFunc("/srs_close", srs_close)
	http.HandleFunc("/srs_publish", srs_publish)
	http.HandleFunc("/srs_unpublish", srs_unpublish)
	http.HandleFunc("/srs_play", srs_play)
	http.HandleFunc("/srs_stop", srs_stop)
	http.HandleFunc("/srs_dvr", srs_dvr)
	if err := http.ListenAndServe("127.0.0.1:10002", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
