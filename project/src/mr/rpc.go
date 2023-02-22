package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
)

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type HeartBeatReq struct {
	WorkerId int
}

type HeartBeatResp struct {
	WorkerId int
}

type AskTaskReq struct {
	WorkerId int
}

type AskTaskResp struct {
	Action    TaskAction
	TaskId    int
	TaskName  string
	TaskType  TaskType
	ReduceCnt int
}

type TaskDoneReq struct {
	WorkerId int
	TaskType TaskType
	TaskId   int
	TaskName string
	Success  bool
}

type TaskDoneResp struct {
}
