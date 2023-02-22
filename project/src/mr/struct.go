package mr

type WorkerState int

const (
	IdleState WorkerState = iota
	BusyState
	DeadState
)

type MasterPhase int

const (
	MapPhase MasterPhase = iota + 1
	ReducePhase
	DonePhase
)

type KeyValue struct {
	Key   string
	Value string
}

const WorkerMaxWaitTime = 10 * 1000 // ms

type TaskType int

const (
	MapTask TaskType = iota + 1
	ReduceTask
)

type TaskAction int

const (
	ActionStayIdle TaskAction = iota + 1
	ActionDoTask
)
