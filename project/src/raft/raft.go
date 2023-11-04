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
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"6.824/labrpc"
)

// import "bytes"
// import "6.824/labgob"

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

type LogEntry struct {
	Index int
	Value interface{}
	Term  int
}

func (e *LogEntry) String() string {
	return fmt.Sprintf("(term: %v, index: %v, value: %+v)", e.Term, e.Index, e.Value)
}

type LogArray struct {
	entries []*LogEntry
	mu      sync.Mutex
}

func GetNewLogArray() *LogArray {
	return &LogArray{
		entries: []*LogEntry{nil},
	}
}

func (l *LogArray) GetNextIndex() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.entries)
}

func (l *LogArray) GetEntry(index int) *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index >= len(l.entries) {
		return nil
	}

	return l.entries[index]
}

func (l *LogArray) Append(value interface{}, term int) *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := &LogEntry{
		Value: value,
		Term:  term,
		Index: len(l.entries),
	}
	l.entries = append(l.entries, entry)
	return entry
}

func (l *LogArray) GetLastLogEntry() *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return nil
	}

	return l.entries[len(l.entries)-1]
}

func (l *LogArray) GetLastNLogEntry(n int) []*LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return nil
	}

	if len(l.entries)-n < 1 {
		return l.entries[1:]
	}

	return l.entries[len(l.entries)-n:]
}

func (l *LogArray) DeleteToIndex(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 || index >= len(l.entries) {
		return
	}

	l.entries = l.entries[:index+1]
	return
}

func (l *LogArray) SetEntries(entries []*LogEntry) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, entry := range entries {
		if entry.Index == len(l.entries) {
			l.entries = append(l.entries, entry)
			continue
		}
		if entry.Index < len(l.entries) {
			entryInLog := l.entries[entry.Index]
			entryInLog.Term = entry.Term
			entryInLog.Value = entry.Value
			continue
		}
		return false
	}
	return true
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

	state               int32
	term                int32
	electionMonitorChan chan bool
	nextElectionTime    int64
	voted               int32
	votedFor            int

	cancelLeader context.CancelFunc

	commitIndex      int32
	lastApplied      int
	triggerApplyChan chan int

	logs        *LogArray
	nextIndexs  sync.Map
	matchIndexs sync.Map

	peerLogEntryTriggerChan map[int]chan bool

	applyCh chan ApplyMsg
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
		peers:                   peers,
		persister:               persister,
		me:                      me,
		state:                   int32(RaftStateFollower),
		term:                    0,
		electionMonitorChan:     make(chan bool, 8),
		nextElectionTime:        getNextElectionTime(),
		voted:                   0,
		nextIndexs:              sync.Map{},
		matchIndexs:             sync.Map{},
		logs:                    GetNewLogArray(),
		applyCh:                 applyCh,
		peerLogEntryTriggerChan: map[int]chan bool{},
		commitIndex:             0,
		lastApplied:             0,
		triggerApplyChan:        make(chan int, 8),
	}

	// Your initialization code here (2A, 2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.electionCheckProcess()
	go rf.applyProcess()

	return rf
}

func (rf *Raft) applyProcess() {
	for {
		if rf.killed() {
			return
		}

		for rf.lastApplied < int(atomic.LoadInt32(&rf.commitIndex)) {
			logEntry := rf.logs.GetEntry(rf.lastApplied + 1)

			DPrintf("server %v begin apply entry: %+v, lastapply: %v, commitIndex: %v",
				rf.me, logEntry, rf.lastApplied, rf.commitIndex)

			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      logEntry.Value,
				CommandIndex: logEntry.Index,
			}
			rf.lastApplied += 1
		}

		select {
		case <-rf.triggerApplyChan:
		case <-time.After(50 * time.Millisecond):
		}
	}
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

	term := int(atomic.LoadInt32(&rf.term))

	if rf.killed() {
		return rf.logs.GetNextIndex(), term, false
	}

	if atomic.LoadInt32(&rf.state) != int32(RaftStateLeader) {
		return rf.logs.GetNextIndex(), term, false
	}

	// Your code here (2B).

	entry := rf.logs.Append(command, term)

	rf.triggerSyncEntry()

	return entry.Index, term, atomic.LoadInt32(&rf.state) == int32(RaftStateLeader)
}

func (rf *Raft) triggerSyncEntry() {
	for peer := range rf.peers {
		rf.triggerPeerSyncEntry(peer)
	}
}

func (rf *Raft) triggerPeerSyncEntry(peer int) {
	triggerChan := rf.peerLogEntryTriggerChan[peer]
	if triggerChan != nil && len(triggerChan) == 0 {
		triggerChan <- true
	}
}

func (rf *Raft) isleader() bool {
	return atomic.LoadInt32(&rf.state) == int32(RaftStateLeader)
}

func (rf *Raft) isfollower() bool {
	return atomic.LoadInt32(&rf.state) == int32(RaftStateFollower)
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	// Your code here (2A).
	return int(atomic.LoadInt32(&rf.term)), rf.isleader()
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

func (rf *Raft) loseLeader() {
	if rf.cancelLeader != nil {
		rf.cancelLeader()
	}
}

func (rf *Raft) changeToFollower(term int) {
	DPrintf("server %v begin chagne to follower, commitindex: %v", rf.me, rf.commitIndex)
	atomic.StoreInt32(&rf.term, int32(term))
	if rf.isleader() {
		rf.loseLeader()
	}
	atomic.StoreInt32(&rf.state, int32(RaftStateFollower))
	DPrintf("server %v change to follower, delete to index: %v", rf.me, rf.commitIndex)
	rf.logs.DeleteToIndex(int(rf.commitIndex))
	rf.resetElectionTimeOut()

}

func (rf *Raft) changeToCandidate() {
	atomic.AddInt32(&rf.term, 1)
	if rf.isleader() {
		rf.loseLeader()
	}
	atomic.StoreInt32(&rf.state, int32(RaftStateCandidate))
	rf.resetElectionTimeOut()
}

func (rf *Raft) changeToLeader() {
	if atomic.LoadInt32(&rf.state) == int32(RaftStateLeader) {
		return
	}
	atomic.StoreInt32(&rf.state, int32(RaftStateLeader))
	DPrintf("server %v become leader", rf.me)

	rf.peerLogEntryTriggerChan = map[int]chan bool{}
	// init nextIndex and matchIndex
	rf.nextIndexs = sync.Map{}
	rf.matchIndexs = sync.Map{}
	nextLogIndex := rf.logs.GetNextIndex()
	for peer := range rf.peers {
		rf.nextIndexs.Store(peer, nextLogIndex)
		rf.matchIndexs.Store(peer, 0)
		rf.peerLogEntryTriggerChan[peer] = make(chan bool, 4)
	}

	rf.resetElectionTimeOut()
	leaderCtx, cancel := context.WithCancel(context.Background())
	rf.cancelLeader = cancel
	go rf.heartBeatProcess(leaderCtx)

	// remove not commit index
	DPrintf("server %v change to leader, delete to index: %v", rf.me, rf.commitIndex)
	rf.logs.DeleteToIndex(int(atomic.LoadInt32(&rf.commitIndex)))

	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		go rf.syncLogEntryToPeerProcess(leaderCtx, peer)
	}

}

func (rf *Raft) syncLogEntryToPeerProcess(leaderCtx context.Context, peer int) {
	triggerChan := rf.peerLogEntryTriggerChan[peer]
	for {
		if rf.killed() {
			return
		}

		rf.syncLogEntryToPeer(leaderCtx, peer)

		select {
		case <-leaderCtx.Done():
			return
		case <-triggerChan:
		case <-time.After(time.Millisecond * time.Duration(LogEntryProcessInterval)):
		}

	}
}

func (rf *Raft) syncLogEntryToPeer(leaderCtx context.Context, peer int) {
	value, ok := rf.matchIndexs.Load(peer)
	if !ok {
		return
	}
	lastMatchIndex := value.(int)
	curIndex := rf.logs.GetNextIndex() - 1

	if lastMatchIndex >= curIndex {
		return
	}

	diffLen := 1
	if lastMatchIndex != 0 {
		diffLen = curIndex - lastMatchIndex
	}

	for {
		// if not leader, return
		select {
		case <-leaderCtx.Done():
			return
		default:
		}

		needAppendEntries := rf.logs.GetLastNLogEntry(diffLen)
		if len(needAppendEntries) == 0 {
			return
		}
		args := &AppendEntriesArgs{
			Term:              int(atomic.LoadInt32(&rf.term)),
			LeaderId:          rf.me,
			LeaderCommitIndex: int(atomic.LoadInt32(&rf.commitIndex)),
			Entries:           needAppendEntries,
		}

		preEntry := rf.logs.GetEntry(needAppendEntries[0].Index - 1)
		if preEntry == nil {
			args.PrevLogIndex = 0
			args.PrevLogTerm = 0
		} else {
			args.PrevLogIndex = preEntry.Index
			args.PrevLogTerm = preEntry.Term
		}

		reply := &AppendEntriesReply{}
		ok := rf.sendAppendEntries(peer, args, reply)
		if !ok {
			return
		}

		if reply.Success {
			lastEntry := needAppendEntries[len(needAppendEntries)-1]

			rf.matchIndexs.Store(peer, lastEntry.Index)
			rf.nextIndexs.Store(peer, lastEntry.Index+1)

			DPrintf("update peer %v matchIndex to %v", peer, lastEntry.Index)

			rf.tryUpdateCommit(lastEntry.Index, peer)
			return
		} else {
			if reply.Term > int(atomic.LoadInt32(&rf.term)) {
				rf.changeToFollower(reply.Term)
				return
			}
		}

		firstSyncEntry := needAppendEntries[0]
		rf.nextIndexs.Store(peer, firstSyncEntry.Index-1)
		diffLen = len(needAppendEntries) * 2
	}
}

func (rf *Raft) tryUpdateCommit(lastSyncIndex int, peer int) {
	if int(atomic.LoadInt32(&rf.commitIndex)) >= lastSyncIndex {
		return
	}

	matchedIndex := []int{}

	for p := range rf.peers {
		if p == rf.me {
			continue
		}
		if index, ok := rf.matchIndexs.Load(p); ok {
			matchedIndex = append(matchedIndex, index.(int))
		}
	}

	sort.Ints(matchedIndex)
	mostMatched := matchedIndex[len(matchedIndex)/2]

	DPrintf("server %v matched index list: %+v, leader commitIndex: %v, mostMatch: %v", rf.me, matchedIndex, rf.commitIndex, mostMatched)

	if mostMatched > int(atomic.LoadInt32(&rf.commitIndex)) {
		atomic.StoreInt32(&rf.commitIndex, int32(mostMatched))
		DPrintf("server %v change commitIndex t0: %v", rf.me, mostMatched)
		if len(rf.triggerApplyChan) == 0 {

			DPrintf("trigger apply, leader commitIndex: %v", rf.commitIndex)

			rf.triggerApplyChan <- int(rf.term)
		}
	}
}

func (rf *Raft) heartBeatProcess(leaderCtx context.Context) {
	for {
		if rf.killed() {
			return
		}

		rf.sendHeartBeat()

		select {
		case <-leaderCtx.Done():
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
		Term:              int(atomic.LoadInt32(&rf.term)),
		LeaderId:          rf.me,
		LeaderCommitIndex: int(atomic.LoadInt32(&rf.commitIndex)),
	}

	if lastLogEntry := rf.logs.GetLastLogEntry(); lastLogEntry != nil {
		args.PrevLogIndex = lastLogEntry.Index
		args.PrevLogTerm = lastLogEntry.Term
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

	rfTerm := int(atomic.LoadInt32(&rf.term))
	if reply.Term > rfTerm {
		rf.changeToFollower(reply.Term)
	}

	if !reply.Success {
		rf.triggerPeerSyncEntry(peer)
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

	DPrintf("perr %v begin election", rf.me)

	rf.changeToCandidate()
	atomic.StoreInt32(&rf.voted, 1)

	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		go rf.Election1(i)
	}
}

func (rf *Raft) Election1(peer int) {
	oldTerm := atomic.LoadInt32(&rf.term)
	args := &RequestVoteArgs{
		Term:        int(atomic.LoadInt32(&rf.term)),
		CandidateId: rf.me,
	}

	if lastLogEntry := rf.logs.GetEntry(int(atomic.LoadInt32(&rf.commitIndex))); lastLogEntry != nil {
		args.LastLogIndex = lastLogEntry.Index
		args.LastLogTerm = lastLogEntry.Term
	}

	reply := &RequestVoteReply{}
	ok := rf.sendRequestVote(peer, args, reply)
	if !ok {
		DPrintf("sendRequestVote to peer %v failed", peer)
		return
	}
	curTerm := atomic.LoadInt32(&rf.term)
	if oldTerm < curTerm {
		DPrintf("term changed, old term: %v, curterm: %v", oldTerm, curTerm)
		return
	}
	if reply.VoteGranted {
		atomic.AddInt32(&rf.voted, 1)
		if !rf.isleader() && int(atomic.LoadInt32(&rf.voted)) > len(rf.peers)/2 {
			rf.changeToLeader()
		}
		return
	}

	if reply.Term > int(curTerm) {
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

	rfTerm := int(atomic.LoadInt32(&rf.term))

	if args.Term <= rfTerm {
		reply.VoteGranted = false
		reply.Term = rfTerm
		return
	}

	lastLogEntry := rf.logs.GetEntry(int(atomic.LoadInt32(&rf.commitIndex)))
	if lastLogEntry != nil && lastLogEntry.Index > args.LastLogIndex {
		DPrintf("server: %v, RequestVote failed, lastLogEntry: %+v", rf.me, lastLogEntry)
		reply.VoteGranted = false
		reply.Term = rfTerm
		return
	}

	reply.VoteGranted = true
	reply.Term = args.Term

	// rf.resetElectionTimeOut()

	atomic.StoreInt32(&rf.term, int32(args.Term))
	atomic.StoreInt32(&rf.state, int32(RaftStateFollower))
	rf.votedFor = args.CandidateId
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	if rf.killed() {
		return
	}

	rfTerm := int(atomic.LoadInt32(&rf.term))
	reply.Success = false

	if args.Term < rfTerm {
		reply.Term = rfTerm
		return
	}

	reply.Term = args.Term

	if args.Term > rfTerm {
		atomic.StoreInt32(&rf.term, int32(args.Term))
		if !rf.isfollower() {
			rf.changeToFollower(args.Term)
			return
		} else {
			rf.logs.DeleteToIndex(int(rf.commitIndex))
		}
	}

	rf.resetElectionTimeOut()

	defer func() {
		if reply.Success {
			rf.tryUpdatePeerCommit(args.LeaderCommitIndex, args.Term)
		}
	}()

	if len(args.Entries) == 0 {
		lastEntry := rf.logs.GetLastLogEntry()
		if lastEntry == nil || lastEntry.Term == args.PrevLogTerm && lastEntry.Index == args.PrevLogIndex {
			reply.Success = true
		}
		return
	}

	if args.PrevLogIndex == 0 {
		rf.logs.DeleteToIndex(0)
		if ok := rf.logs.SetEntries(args.Entries); !ok {
			reply.Success = false
		} else {
			reply.Success = true
		}
		return
	}

	prevEntry := rf.logs.GetEntry(args.PrevLogIndex)
	if prevEntry == nil || prevEntry.Term != args.PrevLogTerm {
		reply.Success = false
		return
	}

	rf.logs.DeleteToIndex(args.PrevLogIndex)
	if ok := rf.logs.SetEntries(args.Entries); !ok {
		reply.Success = false
		return
	}

	reply.Success = true
}

func (rf *Raft) tryUpdatePeerCommit(commitIndex int, term int) {
	lastEntry := rf.logs.GetLastLogEntry()

	if lastEntry == nil {
		return
	}

	// if lastEntry.Term != term {
	// 	return
	// }

	if lastEntry.Index <= commitIndex {
		commitIndex = lastEntry.Index
	}

	if commitIndex > int(atomic.LoadInt32(&rf.commitIndex)) {
		atomic.StoreInt32(&rf.commitIndex, int32(commitIndex))
		DPrintf("tryUpdatePeerCommit server %v change commitIndex t0: %v", rf.me, commitIndex)
		if len(rf.triggerApplyChan) == 0 {
			rf.triggerApplyChan <- term
		}
	}
}
