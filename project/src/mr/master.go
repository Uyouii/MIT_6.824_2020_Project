package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"6.824/common"
)

type Machine struct {
	Id                 int
	State              WorkerState
	HeartBeatTimeStamp int64
	CurrentTask        *MachineTask
}

type MachineTask struct {
	id             int
	name           string
	workerId       int
	done           bool
	taskType       TaskType
	beginTimeStamp int64
}

type Master struct {
	workers  sync.Map
	curId    int
	genIdMux sync.Mutex
	server   *Server
	phase    MasterPhase

	mapTasks     sync.Map // key id, value Task
	undoMapTasks []int
	undoMapMux   sync.Mutex
	reduceTasks  sync.Map // key id, value Task

	undoReduceTasks []int
	undoReduceMux   sync.Mutex

	reduceCnt int
	mapCnt    int
}

type Server struct {
	master *Master
}

func (m *Master) StartServer() {
	m.server = &Server{
		master: m,
	}
	rpc.Register(m.server)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (m *Master) Done() bool {
	return m.phase == DonePhase
}

func MakeMaster(files []string, nReduce int) *Master {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	m := &Master{
		curId:     0,
		phase:     MapPhase,
		reduceCnt: nReduce,
		mapCnt:    len(files),
	}

	// init map tasks
	for index, file := range files {
		task := &MachineTask{
			id:       index,
			name:     file,
			workerId: 0,
			done:     false,
			taskType: MapTask,
		}
		m.mapTasks.Store(task.id, task)
		m.undoMapTasks = append(m.undoMapTasks, task.id)
	}

	// init reduce tasks
	for i := 0; i < nReduce; i++ {
		task := &MachineTask{
			id:       i,
			workerId: 0,
			done:     false,
			taskType: ReduceTask,
		}
		m.reduceTasks.Store(task.id, task)
		m.undoReduceTasks = append(m.undoReduceTasks, task.id)
	}

	m.StartServer()
	// loop sentinel check whether task failed
	go loop(m.Sentinel, time.Second)
	log.Printf("master started...")
	return m
}

func (m *Master) GenWorkerId() int {
	m.genIdMux.Lock()
	defer m.genIdMux.Unlock()
	m.curId++
	return m.curId
}

// rpc...

func (s *Server) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (s *Server) HeartBeat(req *HeartBeatReq, resp *HeartBeatResp) error {
	master := s.master
	if DEBUG {
		log.Printf("reveive hear beat req: %+v", req)
	}
	var worker *Machine
	if req.WorkerId == 0 {
		worker = &Machine{
			Id:                 master.GenWorkerId(),
			State:              IdleState,
			HeartBeatTimeStamp: GetMsTime(),
		}
		master.workers.Store(worker.Id, worker)
	} else {
		v, ok := master.workers.Load(req.WorkerId)
		if !ok {
			log.Fatalf("somethine wrong, can find't worker id: %v in master", req.WorkerId)
		}
		worker = v.(*Machine)
		worker.HeartBeatTimeStamp = GetMsTime()
	}
	resp.WorkerId = worker.Id
	return nil
}

func (m *Master) CheckMapTaskDone() bool {
	if len(m.undoMapTasks) > 0 {
		return false
	}
	for i := 0; i < m.mapCnt; i++ {
		v, _ := m.mapTasks.Load(i)
		task := v.(*MachineTask)
		if !task.done {
			return false
		}
	}
	return true
}

func (m *Master) CheckReduceTaskDone() bool {
	if len(m.undoReduceTasks) > 0 {
		return false
	}
	for i := 0; i < m.reduceCnt; i++ {
		v, _ := m.reduceTasks.Load(i)
		task := v.(*MachineTask)
		if !task.done {
			return false
		}
	}
	return true
}

func (m *Master) AssignMapTask(worker *Machine, resp *AskTaskResp) {

	if len(m.undoMapTasks) == 0 {
		return
	}

	for len(m.undoMapTasks) > 0 {
		taskId := PopFromIntList(&m.undoMapTasks, &m.undoMapMux)
		v, ok := m.mapTasks.Load(taskId)
		if !ok {
			log.Fatalf("somethine wrong, can load task from tasks")
			return
		}
		task := v.(*MachineTask)
		if task.done {
			if DEBUG {
				log.Printf("map task already done: %+v", task)
			}
			continue
		}

		task.workerId = worker.Id
		task.beginTimeStamp = GetMsTime()

		worker.State = BusyState
		worker.CurrentTask = task

		resp.Action = ActionDoTask
		resp.TaskId = task.id
		resp.TaskName = task.name
		resp.ReduceCnt = m.reduceCnt
		resp.TaskType = MapTask
		return
	}

	if m.CheckMapTaskDone() {
		m.phase = ReducePhase
		log.Printf("all map task done, change to reduce phase")
	}
}

func (m *Master) AssignReduceTask(worker *Machine, resp *AskTaskResp) {

	if len(m.undoReduceTasks) == 0 {
		return
	}

	for len(m.undoReduceTasks) > 0 {
		taskId := PopFromIntList(&m.undoReduceTasks, &m.undoReduceMux)
		v, ok := m.reduceTasks.Load(taskId)
		if !ok {
			log.Fatalf("somethine wrong, can load task from tasks")
			return
		}
		task := v.(*MachineTask)
		if task.done {
			log.Printf("reduce task already done: %+v", task)
			continue
		}

		task.workerId = worker.Id
		task.beginTimeStamp = GetMsTime()

		worker.State = BusyState
		worker.CurrentTask = task

		resp.Action = ActionDoTask
		resp.TaskId = task.id
		resp.TaskType = ReduceTask
		resp.MapCnt = m.mapCnt
		return
	}
	if m.CheckReduceTaskDone() {
		log.Printf("all reduce task done, change to done phase")
		m.phase = DonePhase
	}
}

func (s *Server) AskTask(req *AskTaskReq, resp *AskTaskResp) error {
	if req.WorkerId == 0 {
		return common.GetErrorWithMsg(common.InvalidWorkerId, fmt.Sprintf("req worker id: %v", req.WorkerId))
	}

	master := s.master

	v, ok := master.workers.Load(req.WorkerId)
	if !ok {
		return common.GetErrorWithMsg(common.InvalidWorkerId, fmt.Sprintf("req worker id: %v, can't find in master", req.WorkerId))
	}
	worker := v.(*Machine)

	resp.Action = ActionStayIdle
	switch master.phase {
	case DonePhase:
		resp.Action = ActionStayIdle
	case MapPhase:
		master.AssignMapTask(worker, resp)
	case ReducePhase:
		master.AssignReduceTask(worker, resp)
	default:
		log.Fatalf("somethine wrong, invalid master phase: %v", master.phase)
		return nil
	}
	return nil
}

func (m *Master) MapTaskDone(worker *Machine, taskData *TaskDoneReq) {
	taskId := taskData.TaskId
	v, ok := m.mapTasks.Load(taskId)
	if !ok {
		log.Fatalf("invalid task id, can't find in master, task: %+v", taskData)
		return
	}
	task := v.(*MachineTask)

	if task.done {
		// task already done, don't need do anything
		return
	}

	// task failed
	if !taskData.Success {
		// if task cur worker id don't match, may be has assign to another worker
		if task.workerId == worker.Id {
			m.ReleaseTask(task)
			log.Printf("map task %v failed", task.id)
		}
		return
	}

	log.Printf("map task %v done by worker: %v", task.id, worker.Id)
	task.done = true
	task.beginTimeStamp = 0
	if m.CheckMapTaskDone() {
		m.phase = ReducePhase
		log.Printf("all map task done, change to reduce phase")
	}
}

func (m *Master) DeleteInterFiles(task *MachineTask) {
	for mapIndex := 0; mapIndex < m.mapCnt; mapIndex++ {
		filename := fmt.Sprintf("mr_%v_%v.txt", mapIndex, task.id)
		os.Remove(filename)
	}
}

func (m *Master) ReduceTaskDone(worker *Machine, taskData *TaskDoneReq) {
	taskId := taskData.TaskId
	v, ok := m.reduceTasks.Load(taskId)
	if !ok {
		log.Fatalf("invalid task id, can't find in master, task: %+v", taskData)
		return
	}
	task := v.(*MachineTask)

	if task.done {
		// task already done, don't need do anything
		return
	}

	// task failed
	if !taskData.Success {
		// if task cur worker id don't match, may be has assign to another worker
		if task.workerId == worker.Id {
			m.ReleaseTask(task)
			log.Printf("reduce task %v failed", task.id)
		}
		return
	}

	log.Printf("reduce task %v done by worker: %v", task.id, worker.Id)
	task.done = true
	task.beginTimeStamp = 0

	m.DeleteInterFiles(task)

	if m.CheckReduceTaskDone() {
		log.Printf("all reduce task done, change to done phase")
		m.phase = DonePhase
	}
}

func (s *Server) TaskDone(req *TaskDoneReq, resp *TaskDoneResp) error {
	if req.WorkerId == 0 || req.TaskType == 0 {
		return common.GetErrorWithMsg(common.InvalidRequest, fmt.Sprintf("req: %+v", req))
	}

	master := s.master

	v, ok := master.workers.Load(req.WorkerId)
	if !ok {
		return common.GetErrorWithMsg(common.InvalidWorkerId, fmt.Sprintf("req worker id: %v, can't find in master", req.WorkerId))
	}
	worker := v.(*Machine)

	worker.CurrentTask = nil
	worker.State = IdleState

	switch req.TaskType {
	case MapTask:
		master.MapTaskDone(worker, req)
	case ReduceTask:
		master.ReduceTaskDone(worker, req)
	default:
		return common.GetErrorWithMsg(common.InvalidTaskType, fmt.Sprintf("task type: %v, req: %+v", req.TaskType, req))
	}
	return nil
}

func (m *Master) ReleaseTask(task *MachineTask) {
	task.workerId = 0
	task.done = false
	switch task.taskType {
	case MapTask:
		PushToIntList(&m.undoMapTasks, &m.undoMapMux, task.id)
	case ReduceTask:
		PushToIntList(&m.undoReduceTasks, &m.undoReduceMux, task.id)
	}
}

// check worker heartbeat, cancel the worker task
func (m *Master) Sentinel() {

	nowTime := GetMsTime()

	workerCheck := func(key, value interface{}) bool {
		worker := value.(*Machine)
		if worker.State == BusyState && nowTime-worker.HeartBeatTimeStamp > WorkerFailMaxHeartBeatTime {
			// worker deaded
			if worker.CurrentTask != nil {
				m.ReleaseTask(worker.CurrentTask)
			}

			worker.State = DeadState
			worker.CurrentTask = nil
		}
		return true
	}

	m.workers.Range(workerCheck)

	taskCheck := func(key, value interface{}) bool {
		task := value.(*MachineTask)
		if task.workerId != 0 && task.done == false && nowTime-task.beginTimeStamp > TaskFailMaxTime {
			// worker may dead, task failed
			m.ReleaseTask(task)
		}
		return true
	}
	m.mapTasks.Range(taskCheck)
	m.reduceTasks.Range(taskCheck)
}
