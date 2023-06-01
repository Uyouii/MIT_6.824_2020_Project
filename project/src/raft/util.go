package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if DEBUG == true {
		log.Printf(format, a...)
	}
	return
}

func getRandomElectionTimeOut() int64 {
	return rand.Int63n(ElectionTimeOutMax-ElectionTimeOutMin) + ElectionTimeOutMin
}

func getNextElectionTime() int64 {
	return time.Now().UnixMilli() + getRandomElectionTimeOut()
}

func lockSet(n *int, mu *sync.Mutex, v int) {
	mu.Lock()
	defer mu.Unlock()
	*n = v
}

func lockGet(n *int, mu *sync.Mutex) int {
	mu.Lock()
	defer mu.Unlock()
	return *n
}

func lockAdd(n *int, mu *sync.Mutex, v int) {
	mu.Lock()
	defer mu.Unlock()
	*n += v
}
