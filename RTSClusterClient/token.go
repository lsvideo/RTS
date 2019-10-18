// token
package main

import (
	"encoding/json"
	"errors"
	"zkhelper"
)

type Token struct {
	uid        string
	cid        string
	Token      string `json:"token"`      // token字符串
	IP         string `json:"ip"`         //IP地址
	length     int    `json:"length"`     //有效时长
	CreateTime int32  `json:"createTime"` //创建时间
}

func (t *Token) GetToken(uid string) bool {
	var res bool = true
	var tokennode zkhelper.ZKNode //cluster manager服务 注册至自动发现节点
	tokennode.SetServiceType(zkhelper.ServiceTypeRTMP)
	tokennode.SetPath(zkhelper.GetNodePath(zkhelper.GetServicePath(tokennode.ServiceType), zkhelper.NodeTypeToken) + "/" + uid)
	tokennode.SetName(uid)

	str, err := rtsclient.Get(&tokennode)
	if err != nil {
		log.Warningln("Get token err:", err)
		res = false
	} else {
		log.Warningln("Get token :", str)
		err = json.Unmarshal([]byte(str), t)
		if err != nil {
			log.Warningln("Parse json token err:", err)
			res = false
		}
	}
	return res
}

func (t *Token) Verify(vt string, ip string) error {

	//if t.Token == vt && t.IP == ip {
	if t.Token == vt && vt != "" {
		//if t.Token == vt {
		return nil
	}
	return errors.New("Token error.")
}
