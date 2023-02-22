package mr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"6.824/common"
)

type Worker struct {
	Id         int
	mapFunc    func(string, string) []KeyValue
	reduceFunc func(string, []string) string
}

func MakeWorker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	worker := &Worker{
		mapFunc:    mapf,
		reduceFunc: reducef,
	}
	worker.Run()
}

func (w *Worker) Run() {
	// begin heartbeat
	go w.HeartBeat()

	for {
		if !w.AskTask() {
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (w *Worker) AskTask() bool {
	if w.Id == 0 {
		return false
	}
	req := &AskTaskReq{
		WorkerId: w.Id,
	}
	resp := &AskTaskResp{}

	res := call("Server.AskTask", req, resp)
	if !res {
		log.Printf("AskTask failed, workerid: %v", w.Id)
		return false
	}

	if resp.Action == ActionStayIdle {
		log.Printf("worker stay idle, workerid: %v", w.Id)
		return false
	}

	log.Printf("get task from master: %+v", resp)

	w.DoTask(resp)

	return true
}

func (w *Worker) HeartBeat() {
	var successTime int64

	do := func() {
		req := &HeartBeatReq{
			WorkerId: w.Id,
		}
		resp := &HeartBeatResp{}

		res := call("Server.HeartBeat", req, resp)
		if !res {
			return
		}

		successTime = GetMsTime()
		if resp.WorkerId != w.Id && resp.WorkerId != 0 {
			w.Id = resp.WorkerId
		}
	}

	for {
		if successTime != 0 && GetMsTime()-successTime >= WorkerMaxWaitTime {
			log.Printf("master may be dead, stop...")
			os.Exit(1)
		}
		do()
		time.Sleep(time.Second)
	}
}

func (w *Worker) DoTask(task *AskTaskResp) error {
	var err error
	switch task.TaskType {
	case MapTask:
		err = w.DoMap(task)
	case ReduceTask:
		err = w.DoReduce(task)
	default:
		err = common.GetErrorWithMsg(common.UnknownTask, fmt.Sprintf("Unknown task type: %v", task.TaskType))
	}

	req := &TaskDoneReq{
		WorkerId: w.Id,
		TaskType: task.TaskType,
		TaskId:   task.TaskId,
		TaskName: task.TaskName,
	}
	resp := &TaskDoneResp{}

	if err != nil {
		log.Printf("DoTask Error: %v, task: %+v", err, task)
		req.Success = false
	} else {
		log.Printf("DoTask Success, task: %+v", task)
		req.Success = true
	}

	call("Server.TaskDone", req, resp)
	return nil
}

func (w *Worker) DoMap(task *AskTaskResp) error {

	if task.ReduceCnt == 0 {
		return common.GetErrorWithMsg(common.InvalidReduceCnt, fmt.Sprintf("reduce cnt: %v", task.ReduceCnt))
	}

	filename := task.TaskName
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()

	reduceDataList := make([][]KeyValue, task.ReduceCnt)

	kvlist := w.mapFunc(filename, string(content))
	for _, kv := range kvlist {
		reduceId := ihash(kv.Key) % task.ReduceCnt
		reduceDataList[reduceId] = append(reduceDataList[reduceId], kv)
	}

	for reduceId, kvlist := range reduceDataList {
		outFileName := fmt.Sprintf("mr_%v_%v.txt", task.TaskId, reduceId)
		ofile, _ := os.Create(outFileName)
		data, _ := json.Marshal(kvlist)
		ofile.Write(data)
	}

	return nil
}

func (w *Worker) DoReduce(task *AskTaskResp) error {
	return nil
}

func CallExample() {
	// declare an argument structure.
	args := ExampleArgs{}
	// fill in the argument(s).
	args.X = 99
	// declare a reply structure.
	reply := ExampleReply{}
	// send the RPC request, wait for the reply.
	call("Server.Example", &args, &reply)
	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}
