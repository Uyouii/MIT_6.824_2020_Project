package raft

import (
	"log"
	"math/rand"
	"time"
)

var DEBUG = false

func DPrintf(format string, a ...interface{}) {
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
