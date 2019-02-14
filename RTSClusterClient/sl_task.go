// sl_task
package main

import (
	//"net"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	MaxWorkerPoolSize int = 100 * CPUnumbers()
	MaxJobQueueSize   int = 1000
)

type Job interface {
	Do() error
}

// define job channel
type JobChan chan Job

// define worker channel
type WorkerChan chan JobChan

var (
	JobQueue   JobChan    = make(JobChan, MaxJobQueueSize)
	WorkerPool WorkerChan = make(WorkerChan, MaxWorkerPoolSize)
	dispatcher Dispatcher
)

type Worker struct {
	JobChannel JobChan
	quit       chan bool
}

type Dispatcher struct {
	Workers []*Worker
	quit    chan bool
}

type Task struct {
	//未导出字段不会被解析，变量需大写导出
	User_id      string `json:"user_id"`      //user ID
	Task_command string `json:"task_command"` //task command
	Task_data    string `json:"task_data"`    //source url
	finish       chan Task_response
}

type Task_response struct {
	//未导出字段不会被解析，变量需大写导出
	User_id            string `json:"user_id"`      //user ID
	Task_command       string `json:"task_command"` //task type add_input/del_input   add_output/del_output
	Task_status        string `json:"status"`       //task status ok/failed
	Task_response_info string `json:"info"`         //response info
}

func (task Task) Do() error {
	log.Warningln("Do the job!!!!!!!!!!!!!!", task.User_id, task)
	switch task.Task_command {
	case "add_channel":
		addChannel(task.User_id)
		break
	case "get_channel":
		//是否已加入已有channel
		getChannel(task.User_id)
		//若未加入获取本次业务的channel
		getChannel(task.Task_data)
		//channel ++
		break
	case "delete_channel":
		//channel --
		delChannel(task.User_id, false)
		break
	default:
		fmt.Printf("unknown cmd!!!\n")
	}

	//生成应答json
	//fmt.Println(task)
	var resp Task_response
	resp.User_id = task.User_id
	resp.Task_command = task.Task_command
	resp.Task_status = "ok"
	resp.Task_response_info = "OK"
	log.Println(resp)

	task.finish <- resp

	return nil
}

func NewWorker() *Worker {
	return &Worker{make(JobChan), make(chan bool)}
}

func (w *Worker) Start() {
	go func() {
		for { // regist current job channel to worker pool
			WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				if err := job.Do(); err != nil {
					log.Println("excute job failed with err: %v", err)
				} // recieve quit event, stop worker
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (d *Dispatcher) Run() {
	log.Infof("%s:%s Start task dispatcher!", APP_NAME, GetFuncName())
	log.Infof("%s:%s Workpool size %d!", APP_NAME, GetFuncName(), MaxWorkerPoolSize)
	//申请足够的空间
	d.Workers = make([]*Worker, MaxWorkerPoolSize+1)
	for i := 0; i < MaxWorkerPoolSize; i++ {
		worker := NewWorker()
		d.Workers = append(d.Workers, worker)
		worker.Start()
	}
	for {
		select {
		case job := <-JobQueue:
			//t, _ := job.(Task)
			log.Infof("%s:%s Task comming %v!", APP_NAME, GetFuncName(), job.(Task))
			go func(job Job) {
				jobChan := <-WorkerPool
				jobChan <- job
			}(job)
			// stop dispatcher
		case <-d.quit:
			return
		}
	}
}

func init() {
	dispatcher.quit = make(chan bool)
	go dispatcher.Run()
}

func AddTask(job Job) {
	JobQueue <- job
}

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
	case "add_channel", "del_channel", "get_channel":
		AddTask(task)
		resp := <-task.finish
		ret, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(ret))
		break
	default:
		log.Printf("unknown cmd!!!\n")
	}

}

func taskmanager() {
	http.HandleFunc("/vserver", vserver)
	if err := http.ListenAndServe("127.0.0.1:10000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
