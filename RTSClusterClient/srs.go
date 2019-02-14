// srs
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type summaries struct {
	Code int          `json:"code"` // 名称
	Date summary_info `json:"data"`
}

type summary_info struct {
	Status  bool         `json:"ok"` // 名称
	Systime int64        `json:"now_ms"`
	Program program_info `json:"self"`   // 该channel中一个用户被使用次数
	System  system_info  `json:"system"` // 该channel中一个用户被使用次数
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

func get_summaries() (sum *summaries, err error) {
	resp, err := http.Post("http://127.0.0.1:1985/api/v1/summaries",
		"application/x-www-form-urlencoded",
		strings.NewReader("name=slty"))
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
