// sl_interface
package main

import (
	//"net"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//定义服务Protocal
type SL_Protocal interface {
	SLProtocalStart()
	SLProtocalStop()
}

type SL_HTTP struct {
}

var HttpProtocal SL_HTTP

func vserver(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	body_str := string(body)
	log.Println("message: " + body_str)
	log.Println("resp: " + r.RemoteAddr)
	var task Task

	if err := json.Unmarshal(body, &task); err == nil {
		log.Println(task)
	} else {
		log.Println(err)
	}
	task.finish = make(chan Task_response)
	switch task.Task_command {
	case "add_channel", "delete_channel", "get_channel", "leave_channel":
		AddTask(task)
		resp := <-task.finish
		ret, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(ret))
		break
	default:
		log.Printf("unknown cmd!!!\n")
		fmt.Fprintln(w, "unknown cmd")
	}

}

func (sl_http SL_HTTP) SLProtocalStart() {
	http.HandleFunc("/vserver", vserver)
	if err := http.ListenAndServe("127.0.0.1:10000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (sl_http SL_HTTP) SLProtocalStop() {
}
