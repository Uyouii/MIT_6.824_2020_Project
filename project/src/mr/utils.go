package mr

import (
	"hash/fnv"
	"sync"
	"time"
)

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func loop(f func(), looptime time.Duration) {
	for {
		f()
		time.Sleep(looptime)
	}
}

func GetMsTime() int64 {
	return time.Now().UnixNano() / 1000000
}

func PopFromIntList(list *[]int, mux *sync.Mutex) int {
	mux.Lock()
	defer mux.Unlock()
	top := (*list)[0]
	*list = (*list)[1:]
	return top
}

func PushToIntList(list *[]int, mux *sync.Mutex, data int) {
	mux.Lock()
	defer mux.Unlock()
	*list = append(*list, data)
}
