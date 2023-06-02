package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"sync"
	"sync/atomic"
	"time"

	"6.824/labrpc"
)

// import "bytes"
// import "../labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	state               RaftState
	term                int
	termMu              sync.Mutex //
	electionMonitorChan chan bool
	nextElectionTime    int64
	voted               int
	votedMu             sync.Mutex //

	voteCandidate int

	heartBeatCloseChan chan bool
}

func (rf *Raft) isleader() bool {
	return rf.state == RaftStateLeader
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	// Your code here (2A).
	return lockGet(&rf.term, &rf.termMu), rf.isleader()
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1

	// Your code here (2B).

	return index, lockGet(&rf.term, &rf.termMu), rf.state == RaftStateLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	DPrintf("rf %v killed", rf.me)
	// Your code here, if desired.
	close(rf.electionMonitorChan)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:               peers,
		persister:           persister,
		me:                  me,
		state:               RaftStateFollower,
		term:                0,
		electionMonitorChan: make(chan bool, 8),
		nextElectionTime:    getNextElectionTime(),
		heartBeatCloseChan:  make(chan bool, 8),
	}

	// Your initialization code here (2A, 2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.electionCheckProcess()

	return rf
}

func (rf *Raft) loseLeader() {
	rf.heartBeatCloseChan <- true
}

func (rf *Raft) changeToFollower(term int) {
	lockSet(&rf.term, &rf.termMu, term)
	if rf.isleader() {
		rf.loseLeader()
	}
	rf.state = RaftStateFollower
	rf.resetElectionTimeOut()
}

func (rf *Raft) changeToCandidate() {
	lockAdd(&rf.term, &rf.termMu, 1)
	if rf.isleader() {
		rf.loseLeader()
	}
	rf.state = RaftStateCandidate
	rf.resetElectionTimeOut()
}

func (rf *Raft) changeToLeader() {
	DPrintf("peer %v become leader", rf.me)
	rf.state = RaftStateLeader
	rf.resetElectionTimeOut()
	go rf.heartBeatProcess()
}

func (rf *Raft) heartBeatProcess() {
	for {
		if rf.killed() {
			return
		}

		rf.sendHeartBeat()

		select {
		case <-rf.heartBeatCloseChan:
			return
		case <-time.After(time.Millisecond * time.Duration(HeartBeatInterval)):
		}

	}
}

func (rf *Raft) sendHeartBeat() {
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		go rf.sendHeartBeat1(peer)
	}

}

func (rf *Raft) sendHeartBeat1(peer int) {
	args := &AppendEntriesArgs{
		Term:     lockGet(&rf.term, &rf.termMu),
		LeaderId: rf.me,
	}
	reply := &AppendEntriesReply{}
	ok := rf.sendAppendEntries(peer, args, reply)
	if !ok {
		return
	}
	if reply.Success {
		return
	}

	if !rf.isleader() {
		return
	}

	rfTerm := lockGet(&rf.term, &rf.termMu)
	if reply.Term > rfTerm {
		rf.changeToFollower(reply.Term)
	}
}

func (rf *Raft) electionCheckProcess() {
	for {
		if rf.killed() {
			DPrintf("peer %v electionCheckProcess killed", rf.me)
			return
		}

		waitTimeMs := rf.nextElectionTime - time.Now().UnixMilli()

		select {
		case <-rf.electionMonitorChan:
		case <-time.After(time.Millisecond * time.Duration(waitTimeMs)):
			// timeout, begin election
			if !rf.isleader() {
				rf.DoElection()
			}
		}
		rf.nextElectionTime = getNextElectionTime()
	}
}

func (rf *Raft) DoElection() {
	if rf.killed() {
		return
	}
	if DEBUG {
		DPrintf("perr %v begin election", rf.me)
	}
	rf.changeToCandidate()
	lockSet(&rf.voted, &rf.votedMu, 1)

	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		go rf.Election1(i)
	}
}

func (rf *Raft) Election1(peer int) {
	oldTerm := lockGet(&rf.term, &rf.termMu)
	args := &RequestVoteArgs{
		Term:        lockGet(&rf.term, &rf.termMu),
		CandidateId: rf.me,
	}
	reply := &RequestVoteReply{}
	ok := rf.sendRequestVote(peer, args, reply)
	if !ok {
		DPrintf("sendRequestVote to peer %v failed", peer)
		return
	}
	curTerm := lockGet(&rf.term, &rf.termMu)
	if oldTerm < curTerm {
		DPrintf("term changed, old term: %v, curterm: %v", oldTerm, curTerm)
		return
	}
	if reply.VoteGranted {
		lockAdd(&rf.voted, &rf.votedMu, 1)
		if !rf.isleader() && lockGet(&rf.voted, &rf.votedMu) > len(rf.peers)/2 {
			rf.changeToLeader()
		}
		return
	}

	if reply.Term > curTerm {
		rf.changeToFollower(reply.Term)
	}
}

func (rf *Raft) resetElectionTimeOut() {
	if rf.killed() {
		return
	}
	rf.electionMonitorChan <- true
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	if rf.killed() {
		return
	}

	rfTerm := lockGet(&rf.term, &rf.termMu)

	if args.Term <= rfTerm {
		reply.VoteGranted = false
		reply.Term = rfTerm
		return
	}
	reply.VoteGranted = true
	reply.Term = args.Term

	rf.resetElectionTimeOut()

	lockSet(&rf.term, &rf.termMu, args.Term)
	rf.state = RaftStateFollower
	rf.voteCandidate = args.CandidateId
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	if rf.killed() {
		return
	}

	rfTerm := lockGet(&rf.term, &rf.termMu)

	if args.Term < rfTerm {
		reply.Success = false
		reply.Term = rfTerm
		return
	}

	reply.Term = args.Term
	reply.Success = true

	if args.Term > rfTerm {
		rf.changeToFollower(args.Term)
		return
	}

	rf.resetElectionTimeOut()
	return

}
