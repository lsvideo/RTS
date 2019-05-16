// sl_echat
package main

import (
	"proto/conn_penetrate"
	"proto/video"

	zmq "github.com/pebbe/zmq4"

	"github.com/golang/protobuf/proto"

	"encoding/json"
	"fmt"

	//	"io"
	"strconv"
)

type SL_ECHAT struct {
}

type RTCStrMessage struct {
	RtcType *string `json:"rtcType,omitempty"`
	RtcData *string `json:"rtcData,omitempty"`
}

var EchatProtocal SL_ECHAT

func init() {
	RegisterCommand("video.GetChannel", EchatGetChannel)
	RegisterCommand("video.GetChannelResp", EchatGetChannelResp)
	RegisterCommand("video.GetToken", EchatGetToken)
	RegisterCommand("video.GetTokenResp", EchatGetTokenResp)
	RegisterCommand("video.StreamStop", EchatStreamStop)
}

func (p SL_ECHAT) SLProtocalStart() {
	socket, _ := zmq.NewSocket(zmq.ROUTER)
	defer socket.Close()
	socket.Bind(config.Video_server)
	// Wait for messages
	for {
		msg1, _ := socket.Recv(0)
		log.Infoln("Received1 ", msg1)
		msg2, _ := socket.Recv(0)
		log.Infoln("Received2 ", string(msg2))
		msg3, _ := socket.Recv(0)
		log.Infoln("Received3 ", string(msg3))
		msg4, _ := socket.Recv(0)
		log.Infoln("Received msg ", string(msg4))
		//msg5, _ := socket.Recv(0)
		//println("Received5 ", string(msg5))

		msg := &conn_penetrate.ClientMessageIn{}
		proto.Unmarshal([]byte(msg4), msg)
		fmt.Println(msg)
		log.Infoln("Received Name:", msg.GetName(), "Msg:", msg.GetMsg())

		//mms := &conn_penetrate.SendMMS{}
		//proto.Unmarshal(msg.GetMsg(), gc)
		//proto.Unmarshal(msg.GetMsg(), mms)
		//fmt.Println(mms)
		//log.Infoln("Received :", mms.GetTitle(), mms.GetContent())
		var task Task
		task.Task_command = msg.GetName()
		task.Task_data = string(msg.GetMsg())
		AddTask(task)
	}
}

func (p SL_ECHAT) SLProtocalStop() {
}

func echatSendMsg(data string) {
	socket, err := zmq.NewSocket(zmq.DEALER)

	log.Infoln("socket err:", err)
	err = socket.Connect(config.Penetrate_server)
	defer socket.Close()
	log.Infoln("Connect err:", err)
	if nil != err {
		log.Errorln(err)
	}

	var c int
	c, err = socket.Send("", zmq.SNDMORE)
	log.Infoln("Send \" \" count:", c, "err:", err)
	if nil != err {
		log.Errorln(err)
	}
	c, err = socket.Send("#pb", zmq.SNDMORE)
	log.Infoln("Send #pb count:", c, "err:", err)
	if nil != err {
		log.Errorln(err)
	}
	c, err = socket.Send("conn.penetrate.Penetrate", zmq.SNDMORE)
	log.Infoln("Send conn.penetrate.Penetrate count:", c, "err:", err)
	if nil != err {
		log.Errorln(err)
	}

	log.Infoln("Send msg :", data)
	//fmt.Println("!!!", pen)
	//	f, _ := os.OpenFile("123", os.O_CREATE|os.O_WRONLY, 0666)
	//	defer f.Close()
	//	f.Write(buf)
	c, err = socket.Send(data, 0)

	log.Infoln("Send msg over count:", c, "err:", err)
	if nil != err {
		log.Errorln(err)
	}
}

func EchatGetChannelResp(t Task) {
	uid64, _ := strconv.ParseUint(t.User_id, 0, 32)
	uid := uint32(uid64)
	log.Infoln("send Resp to :", uid)
	url := getMinLinksCahnnel()
	log.Warningln("!!!!!!!!!!!!!url :", url)
	buf := PenetrateRTSGetChannelResp(uid, url)

	echatSendMsg(string(buf))

	//time.Sleep(1 * time.Second)
}

func EchatGetChannel(t Task) {
	gc := &video.GetChannel{}
	json.Unmarshal([]byte(t.Task_data), gc)
	var task Task
	task.Task_command = "video.GetChannelResp"
	task.User_id = gc.GetUid()
	AddTask(task)
}

func PenetrateRTSGetChannelResp(uid uint32, url string) []byte {
	pen := &conn_penetrate.Penetrate{}
	msg := &conn_penetrate.UserMessage{}
	external := &conn_penetrate.ExternalServMsg{}
	msgexternal := &video.RTCMessage{}
	chn := &video.Item{}
	//构造GetChannelResp
	chn.Key = proto.String("url")
	chn.Value = proto.String(url)

	//构造填充在ExternalServMsg中的UserMessage
	msgexternal.RtcType = proto.String("video.GetChannelResp")
	msgexternal.RtcArrayData = append(msgexternal.RtcArrayData, chn)
	//bufmsgexternal, _ := proto.Marshal(msgexternal) //序列化
	jsonStu, _ := json.Marshal(msgexternal)

	log.Printf("byte: %v \n", jsonStu)
	log.Printf("String: %s\n", string(jsonStu))
	//构造ExternalServMsg

	//external.Msg = proto.String(string(bufmsgexternal))
	external.Msg = proto.String(string(jsonStu))
	//external.Msg = proto.String("test")
	bufferext, _ := proto.Marshal(external) //序列化

	msg.Name = proto.String("ptt.serv.ExternalServMsg")
	msg.Msg = bufferext

	pen.Msg = append(pen.Msg, msg)
	pen.Tartype = conn_penetrate.Penetrate_USER.Enum()
	pen.Targets = proto.Uint32(uid)

	fmt.Println(pen)
	buffer, _ := proto.Marshal(pen)

	return buffer
}

func PenetrateRTSGetTokenResp(uid uint32, date string) []byte {
	gt := &video.GetToken{}
	json.Unmarshal([]byte(date), gt)

	pen := &conn_penetrate.Penetrate{}
	msg := &conn_penetrate.UserMessage{}
	external := &conn_penetrate.ExternalServMsg{}
	msgexternal := &video.RTCMessage{}

	//构造GetTokenResp
	user := &video.Item{}
	user.Key = proto.String("touid")
	user.Value = proto.String(gt.GetTouid())
	msgexternal.RtcArrayData = append(msgexternal.RtcArrayData, user)
	action := &video.Item{}
	action.Key = proto.String("action")
	action.Value = proto.String(gt.GetAction())
	msgexternal.RtcArrayData = append(msgexternal.RtcArrayData, action)
	token := &video.Item{}
	token.Key = proto.String("token")
	token.Value = proto.String(GetStringMd5(date))
	msgexternal.RtcArrayData = append(msgexternal.RtcArrayData, token)

	//构造填充在ExternalServMsg中的UserMessage
	msgexternal.RtcType = proto.String("video.GetTokenResp")

	//bufmsgexternal, _ := proto.Marshal(msgexternal) //序列化
	jsonStu, _ := json.Marshal(msgexternal)

	log.Printf("byte: %v \n", jsonStu)
	log.Printf("String: %s\n", string(jsonStu))
	//构造ExternalServMsg

	//external.Msg = proto.String(string(bufmsgexternal))
	external.Msg = proto.String(string(jsonStu))
	//external.Msg = proto.String("test")
	bufferext, _ := proto.Marshal(external) //序列化

	msg.Name = proto.String("ptt.serv.ExternalServMsg")
	msg.Msg = bufferext

	pen.Msg = append(pen.Msg, msg)
	pen.Tartype = conn_penetrate.Penetrate_USER.Enum()
	pen.Targets = proto.Uint32(uid)

	fmt.Println(pen)
	buffer, _ := proto.Marshal(pen)

	return buffer
}

func EchatGetToken(t Task) {
	gt := &video.GetToken{}
	json.Unmarshal([]byte(t.Task_data), gt)
	log.Println("####video.GetTokenResp:", t)
	var task Task
	task.Task_command = "video.GetTokenResp"
	task.User_id = gt.GetUid()
	task.Task_data = t.Task_data
	AddTask(task)
}

func EchatGetTokenResp(t Task) {
	uid64, _ := strconv.ParseUint(t.User_id, 0, 32)
	uid := uint32(uid64)
	log.Infoln("send Resp to :", uid)

	buf := PenetrateRTSGetTokenResp(uid, t.Task_data)

	echatSendMsg(string(buf))
}

func PenetrateStreamStopResp(uid uint32, cmd string, data string) []byte {
	pen := &conn_penetrate.Penetrate{}
	msg := &conn_penetrate.UserMessage{}
	external := &conn_penetrate.ExternalServMsg{}
	msgexternal := &video.RTCMessage{}
	chn := &video.Item{}
	//构造GetChannelResp
	var echatchannel eChatChannel
	json.Unmarshal([]byte(data), &echatchannel)
	chn.Key = proto.String("uid")
	chn.Value = proto.String(echatchannel.Uid)

	//构造填充在ExternalServMsg中的UserMessage
	msgexternal.RtcType = proto.String(cmd)
	msgexternal.RtcArrayData = append(msgexternal.RtcArrayData, chn)
	//bufmsgexternal, _ := proto.Marshal(msgexternal) //序列化
	jsonStu, _ := json.Marshal(msgexternal)

	log.Printf("byte: %v \n", jsonStu)
	log.Printf("String: %s\n", string(jsonStu))
	//构造ExternalServMsg

	//external.Msg = proto.String(string(bufmsgexternal))
	external.Msg = proto.String(string(jsonStu))
	//external.Msg = proto.String("test")
	bufferext, _ := proto.Marshal(external) //序列化

	msg.Name = proto.String("ptt.serv.ExternalServMsg")
	msg.Msg = bufferext

	pen.Msg = append(pen.Msg, msg)
	pen.Tartype = conn_penetrate.Penetrate_USER.Enum()
	pen.Targets = proto.Uint32(uid)

	fmt.Println(pen)
	buffer, _ := proto.Marshal(pen)

	return buffer
}

func EchatStreamStop(t Task) {
	uid64, _ := strconv.ParseUint(t.User_id, 0, 32)
	uid := uint32(uid64)
	log.Infoln("send Resp to :", uid)

	buf := PenetrateStreamStopResp(uid, t.Task_command, t.Task_data)

	echatSendMsg(string(buf))
}
