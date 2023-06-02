package raft

type RaftState int

const (
	RaftStateFollower RaftState = iota + 1
	RaftStateCandidate
	RaftStateLeader
)

const (
	ElectionTimeOutMin = 300 // ms
	ElectionTimeOutMax = 600 // ms

	HeartBeatInterval = 100 //ms
)
