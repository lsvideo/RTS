// channel
package main

import (
	"container/list"
	"zkhelper"
)

type VChannel struct {
	Name        string      `json:"name"` // 名称
	ChannelType ChannelType `json:"type"`
	Num         int         `json:"num"` // 该用户channel被使用次数
	lstUsers    *list.List  //用户列表
}

type ChannelType int32

const (
	ChannelTypeUnknown ChannelType = 0
	ChannelTypeRTMP    ChannelType = 1
)

var (
	channelTypeNames = map[ChannelType]string{
		ChannelTypeUnknown: "ChannelTypeUnknown",
		ChannelTypeRTMP:    "ChannelTypeRTMP",
	}
)

func NewChannel() *VChannel {
	return &VChannel{"", ChannelTypeUnknown, 0, list.New()}
}

func getServerLinks(serverid string) int {
	var num int = 0
	value, ok := mapServers.Load(serverid)
	server := value.(VServer)
	if ok {
		num += server.ChannelNum
		for e := server.lstChannels.Front(); e != nil; e = e.Next() {
			user := e.Value.(string)
			value, ok := mapChannels.Load(user)
			channel := value.(*VChannel)
			if ok {
				num += channel.Num
			}
		}
		return num
	}
	return -1
}

func addChanneltoServer(serverid string, userid string) {
	value, ok := mapServers.Load(serverid)
	if ok {
		server := value.(VServer)

		var bExist bool = false
		for e := server.lstChannels.Front(); e != nil; e = e.Next() {
			user := e.Value.(string)
			if user == userid {
				bExist = true
			}
		}
		if !bExist {
			server.lstChannels.PushBack(userid)
			server.ChannelNum++
			mapServers.Store(serverid, server)
		}
	}
}

func addChannel(usrid string) {
	var channel *VChannel
	value, ok := mapChannels.Load(usrid)
	if ok {
		channel = value.(*VChannel)
	} else {
		var minLinks int = -1
		var channelname string = ""

		getMinLinksServer := func(k, v interface{}) bool {
			//这个函数的入参、出参的类型都已经固定，不能修改
			//可以在函数体内编写自己的代码，调用map中的k,v
			server := v.(VServer)
			key := k.(string)
			if server.statenode.Status == zkhelper.NodeStatusWatched {
				links := getServerLinks(key)
				if (minLinks == -1 || links < minLinks) && links != -1 {
					minLinks = links
					channelname = key
					if minLinks == 0 {
						goto finish
					}
				}
			}
			return true
		finish:
			return false
		}
		mapServers.Range(getMinLinksServer)
		if len(channelname) != 0 {
			log.Println("get channel@@@@@@", channelname)
			addChanneltoServer(channelname, usrid)
			channel = NewChannel()
			channel.Name = channelname
		}
		if channel != nil {
			mapChannels.Store(usrid, channel)
			log.Println("@@@@@Channels", mapChannels)
		}
	}

	//mapChannels.Range(sc)
	mapServers.Range(MapGoThrough)
}

func UserInChannel(channel *VChannel, userid string) bool {

	var bExist bool = false
	for e := channel.lstUsers.Front(); e != nil; e = e.Next() {
		user := e.Value.(string)
		if user == userid {
			bExist = true
		}
	}
	return bExist
}

func getChannel(userid string, dstid string) (*VChannel, bool) {
	value, ok := mapChannels.Load(dstid)
	if ok {
		channel := value.(*VChannel)
		if !UserInChannel(channel, userid) {
			channel.Num++
			channel.lstUsers.PushBack(userid)
			mapChannels.Store(dstid, channel)
		}
		mapChannels.Range(MapGoThrough)
		return channel, ok
	}
	return nil, ok
}

func leaveChannel(userid string, dstid string) bool {
	value, ok := mapChannels.Load(dstid)
	if ok {
		channel := value.(*VChannel)
		if UserInChannel(channel, userid) {
			channel.Num--
			delUserfromChannel(channel, userid)
			mapChannels.Store(dstid, channel)
		}
		mapChannels.Range(MapGoThrough)
	}
	return ok
}

func ChannelInServer(server *VServer, channelid string) bool {

	var bExist bool = false
	for e := server.lstChannels.Front(); e != nil; e = e.Next() {
		channel := e.Value.(string)
		if channel == channelid {
			bExist = true
		}
	}
	return bExist
}

func delChannelfromServer(serverid string, channelid string) {
	value, ok := mapServers.Load(serverid)
	if ok {
		server := value.(VServer)
		if ChannelInServer(&server, channelid) {
			server.ChannelNum--
			for e := server.lstChannels.Front(); e != nil; e = e.Next() {
				channel := e.Value.(string)
				if channel == channelid {
					server.lstChannels.Remove(e)
					break
				}
			}
			mapServers.Store(serverid, server)
		}
	}
}

func delUserfromChannel(channel *VChannel, userid string) {
	var next *list.Element
	for e := channel.lstUsers.Front(); e != nil; e = next {
		user := e.Value.(string)
		next = e.Next()
		if user == userid {
			channel.lstUsers.Remove(e)
			break
		} else if "all" == userid {
			channel.lstUsers.Remove(e)
		}
	}
}

func delChannel(userid string) {
	var channel *VChannel
	value, ok := mapChannels.Load(userid)
	if ok {
		channel = value.(*VChannel)
		delChannelfromServer(channel.Name, userid)
		delUserfromChannel(channel, "all")
		mapChannels.Delete(userid)
	}
	mapChannels.Range(MapGoThrough)

}

func CleanServerChannels(serverid string) {
	value, ok := mapServers.Load(serverid)
	if ok {
		server := value.(*VServer)
		var next *list.Element
		for e := server.lstChannels.Front(); e != nil; e = next {
			user := e.Value.(string)
			delChannel(user)
			next = e.Next()
			server.lstChannels.Remove(e)
		}
	}
}
