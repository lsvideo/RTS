// srs
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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

func srs_connect(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)

	w.WriteHeader(200)
	w.Write([]byte("0"))

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
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	w.WriteHeader(200)
	w.Write([]byte("0"))
}

func srs_unpublish(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	w.WriteHeader(200)
	w.Write([]byte("0"))
}

func srs_play(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	w.WriteHeader(200)
	w.Write([]byte("0"))
}

func srs_stop(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	w.WriteHeader(200)
	w.Write([]byte("0"))
}

func srsmanager() {
	http.HandleFunc("/srs_connect", srs_connect)
	http.HandleFunc("/srs_close", srs_close)
	http.HandleFunc("/srs_publish", srs_publish)
	http.HandleFunc("/srs_unpublish", srs_unpublish)
	http.HandleFunc("/srs_play", srs_play)
	http.HandleFunc("/srs_stop", srs_stop)
	if err := http.ListenAndServe("127.0.0.1:10002", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
