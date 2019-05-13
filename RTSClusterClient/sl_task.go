// sl_task
package main

import (
	//"net"
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	//"net/http"
)

var (
	MaxWorkerPoolSize int = CPUnumbers()
	MaxJobQueueSize   int = CPUnumbers()
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
	var resp Task_response
	resp.User_id = task.User_id
	resp.Task_command = task.Task_command

	switch task.Task_command {
	case "add_channel":
		addChannel(task.User_id)
		resp.Task_status = "ok"
		resp.Task_response_info = "OK"
		log.Println(resp)
		task.finish <- resp
		break
	case "get_channel":
		channel, ok := getChannel(task.User_id, task.Task_data)
		if ok {
			resp.Task_status = "ok"
			resp.Task_response_info = channel.Name
		} else {
			resp.Task_status = "error"
			resp.Task_response_info = "get no channel"
		}
		log.Println(resp)
		task.finish <- resp
		break
	case "leave_channel":
		leaveChannel(task.User_id, task.Task_data)
		resp.Task_status = "ok"
		resp.Task_response_info = "OK"
		log.Println(resp)
		task.finish <- resp
	case "delete_channel":
		delChannel(task.User_id)
		resp.Task_status = "ok"
		resp.Task_response_info = "ok"
		log.Println(resp)
		task.finish <- resp
		break
	case "video.GetChannel":
		RunCommand("video.GetChannel", task)
		break
	case "video.GetChannelResp":
		RunCommand("video.GetChannelResp", task)
		break
	case "video.GetToken":
		RunCommand("video.GetToken", task)
		break
	case "video.GetTokenResp":
		RunCommand("video.GetTokenResp", task)
		break
	default:
		fmt.Printf("unknown cmd!!!\n")
	}
	log.Warningln("Do the response!!!!!!!!!!!!!!", task.User_id, task)
	//生成应答json
	//fmt.Println(task)

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
