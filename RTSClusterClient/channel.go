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
	server := value.(*VServer)
	if ok {
		num += server.ChannelNum
		for e := server.lstUsers.Front(); e != nil; e.Next() {
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

func addUserChannel(serverid string, usrid string) {
	value, ok := mapServers.Load(serverid)
	if ok {
		server := value.(*VServer)

		var bExist bool = false
		for e := server.lstUsers.Front(); e != nil; e.Next() {
			user := e.Value.(string)
			if user == userid {
				bExist = true
			}
		}
		if !bExist {
			server.lstUsers.PushBack(usrid)
			server.ChannelNum++
			mapServers.Store(serverid, server)
		}
	}
}

func addChannel(usrid string) {
	channel, ok := getChannel(usrid)
	if !ok {
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
		finish:
			if len(channelname) != 0 {
				log.Println("!!!!!!!", server)
				addUserChannel(channelname, usrid)

			}
			return true
		}
		mapServers.Range(getMinLinksServer)
		if len(channelname) != 0 {
			channel = NewChannel()
			channel.Name = channelname
		}
		log.Println(mapServers)
	}
	if channel != nil {
		mapChannels.Store(usrid, channel)
	}
	log.Println("########", channel)
}

func getChannel(usrid string) (*VChannel, bool) {
	value, ok := mapChannels.Load(usrid)
	if ok {
		channel := value.(*VChannel)
		channel.Num++
		mapChannels.Store(usrid, channel)
		return channel, ok
	}
	return nil, ok
}

func delChannel(usrid string, force bool) {
	channel, ok := getChannel(usrid)
	if ok {
		if force {
			mapChannels.Delete(usrid)
		} else {
			channel.Num--
			if channel.Num == 0 {
				mapChannels.Delete(usrid)
			} else {
				mapChannels.Store(usrid, channel)
			}
		}
	}
}

func CleanServerChannels(serverid string) {
	value, ok := mapServers.Load(serverid)
	if ok {
		server := value.(*VServer)
		var next *list.Element
		for e := server.lstUsers.Front(); e != nil; e = next {
			user := e.Value.(string)
			delChannel(user, true)
			next = e.Next()
			server.lstUsers.Remove(e)
		}
	}
}
