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
	//	"bytes"

	"bytes"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"

	//"runtime" // 5.6添加使用go程立即执行

	"6.5840/labgob"
	"6.5840/labrpc"
)

const (
	STATE_FOLLOWER = iota // 连续递增的常量计数器
	STATE_CANDIDATE
	STATE_LEADER
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.

// 每次向日志提交新条目时，每个Raft peer should send an ApplyMsg to the service (or tester).
type ApplyMsg struct {
	CommandValid bool        // 操作有效性
	Command      interface{} // 操作
	CommandIndex int         // 操作所在的index

	// For 2D:
	SnapshotValid bool   // 快照的有效性
	Snapshot      []byte // 快照
	SnapshotTerm  int    // 快照对应最后一个log的term
	SnapshotIndex int    // 快照对应最后一个log的index
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // to protect shared acLockcess to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	commitIndex int   // 已经commit的最后一个log的index，每个节点的值不一样，但是最后趋于相同
	lastApplied int   // 已经appliy到状态机的最后一个log的index,也就是被执行的指令，每个节点相同
	nextIndex   []int // leader认为的所有server下次需要新得到的log的index,乐观值
	matchIndex  []int // leader保证的所有server的log一致的最后一个，悲观值

	state       uint32  // 0 follower 1 candidata 2 leader
	logs        []Entry // 日志数组
	currentTerm int     // 当前任期
	voteFor     int     // 当前任期收到选票的candidateId

	electionTime  time.Time     // 下一次选举的时间点
	heartbeatTime time.Duration // 心跳时间间隔

	applyCh    chan ApplyMsg // 提交操作的应用通道
	applyVCond *sync.Cond    // 提交操作的通道使用的条件变量

	//2D
	LastIncludedIndex int // 快照保存的最后一个log的index
	LastIncludedTerm  int // 快照保存的最后一个log的term
}

type Entry struct {
	Index int
	Term  int
	Cmd   interface{} // 具体的操作
}

// return currentTerm and whether this server
// believes it is the leader.
// 上层或者测试 ask a Raft for its current term, and whether it thinks it is leader
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool

	rf.mu.Lock()
	defer rf.mu.Unlock()
	// Your code here (2A).
	term = int(rf.currentTerm)
	isleader = rf.state == STATE_LEADER
	return term, isleader
}

//更新raft的状态
func (rf *Raft) UpdateState(state uint32) {
	if rf.state == state {
		return
	}
	////old_state := rf.state
	rf.state = state
	////log.Printf("{machine %v} update state from %v to %v, term %v", rf.me, old_state, state, rf.currentTerm)
}

// 得到raft最后一个log
func (rf *Raft) GetlastLog() Entry {
	return rf.logs[len(rf.logs)-1]
}

// 得到raft最新的log的index
func (rf *Raft) GetRealLastLogIndex() int {
	return Max(rf.logs[len(rf.logs)-1].Index, rf.LastIncludedIndex)
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) encoderaftstate() []byte {
	newBuf := new(bytes.Buffer)
	newEncoder := labgob.NewEncoder(newBuf)
	err := newEncoder.Encode(rf.voteFor)
	if err != nil {
		log.Printf("{machine[%d]}.state[%v].term[%d]: encode voteFor error: %v\n", rf.me, rf.state, rf.currentTerm, err)
	}
	err = newEncoder.Encode(rf.currentTerm)
	if err != nil {
		log.Printf("{machine[%d]}.state[%v].term[%d]: encode currentTerm error: %v\n", rf.me, rf.state, rf.currentTerm, err)
	}
	err = newEncoder.Encode(rf.logs)
	if err != nil {
		log.Printf("{machine[%d]}.state[%v].term[%d]: encode logs error: %v\n", rf.me, rf.state, rf.currentTerm, err)
	}
	err = newEncoder.Encode(rf.LastIncludedIndex)
	if err != nil {
		log.Printf("{machine[%d]}.state[%v].term[%d]: encode logs error: %v\n", rf.me, rf.state, rf.currentTerm, err)
	}
	err = newEncoder.Encode(rf.LastIncludedTerm)
	if err != nil {
		log.Printf("{machine[%d]}.state[%v].term[%d]: encode logs error: %v\n", rf.me, rf.state, rf.currentTerm, err)
	}

	raftstate := newBuf.Bytes()
	return raftstate
}

func (rf *Raft) persistState() {
	//Your code here (2C).
	rf.persister.SaveStateOnly(rf.encoderaftstate())
	////log.Printf("{machine %v }persist state", rf.me)
}

func (rf *Raft) persistStateAndSnapshot(snapshot []byte) {
	rf.persister.Save(rf.encoderaftstate(), snapshot)
	////log.Printf("{machine %v }persist state and snapshot", rf.me)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state? 空指针和空值
		return
	}

	newBuf := bytes.NewBuffer(data)
	newDecoder := labgob.NewDecoder(newBuf)
	var old_voteFor int
	var old_currentTerm int
	var old_logs []Entry
	var old_lastincludedindex int
	var old_lastincludedterm int

	rf.mu.Lock()
	defer rf.mu.Unlock()

	if newDecoder.Decode(&old_voteFor) != nil ||
		newDecoder.Decode(&old_currentTerm) != nil ||
		newDecoder.Decode(&old_logs) != nil ||
		newDecoder.Decode(&old_lastincludedindex) != nil ||
		newDecoder.Decode(&old_lastincludedterm) != nil {
		// 无法解码
		log.Printf("{machine[%d]}.state[%v].term[%d]: Decode error", rf.me, rf.state, rf.currentTerm)
	} else {
		rf.voteFor = old_voteFor
		rf.logs = old_logs
		rf.currentTerm = old_currentTerm
		rf.LastIncludedIndex = old_lastincludedindex
		rf.LastIncludedTerm = old_lastincludedterm
		////log.Printf("{machine %v} restore, voteFor: %v, term: %v, logs: %v，LastIncludedIndex：%v, LastIncludedTerm: %v", rf.me, rf.voteFor, rf.currentTerm, rf.logs, rf.LastIncludedIndex, rf.LastIncludedTerm)
	}

}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

	rf.mu.Lock()
	defer rf.mu.Unlock()

	////log.Printf("start snapshot/////////{machine %v} term{%v} snapshot,LastIncludedIndex: %v,LastIncludedTerm: %v, logslen: %v",
	////	rf.me, rf.currentTerm, rf.LastIncludedIndex, rf.LastIncludedTerm, len(rf.logs))

	if index > rf.GetlastLog().Index || index > rf.lastApplied || index <= rf.LastIncludedIndex {
		return
	}

	rf.LastIncludedTerm = rf.logs[index-rf.LastIncludedIndex].Term //选举的时候用
	rf.logs = append(rf.logs[:1], rf.logs[index+1-rf.LastIncludedIndex:]...)
	rf.LastIncludedIndex = index
	rf.logs[0].Term = rf.LastIncludedTerm

	rf.persistStateAndSnapshot(snapshot)
	////log.Printf("end snapshot /////////{machine %v} term{%v} snapshot,LastIncludedIndex: %v,LastIncludedTerm: %v, logslen: %v",
	////	rf.me, rf.currentTerm, rf.LastIncludedIndex, rf.LastIncludedTerm, len(rf.logs))
}

// InstallSnapshot RPC
type InstallSnapshotArgs struct {
	Term              int
	LeaderId          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte // 快照
	//Offset            int
	//Done              bool
}
type InstallSnapshotReply struct {
	Term    int
	Success bool // 快照应用情况
	Xindex  int  // 期待下次发过来的log的index
}

func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	////log.Printf("{machine %v} self LastIncludedIndex[%v],get snapshot, args.LastIncludedIndex: %v, args.LastIncludedTerm: %v",
	////	rf.me, rf.LastIncludedIndex, args.LastIncludedIndex, args.LastIncludedTerm)

	reply.Term, reply.Success, reply.Xindex = rf.currentTerm, false, rf.GetRealLastLogIndex()+1
	// 判断RPC过时
	if args.Term < rf.currentTerm {
		////log.Printf("{machine %v} InstallSnapshot term out, args.Term < rf.currentTerm", rf.me)
		return
	}

	//servers 2
	if args.Term > rf.currentTerm {
		rf.UpdateState(STATE_FOLLOWER)
		rf.currentTerm = args.Term
		rf.persistState()
	}

	rf.resetElectionTime()

	if rf.state == STATE_CANDIDATE {
		rf.UpdateState(STATE_FOLLOWER)
	}

	// 判断快照有效性
	//if args.LastIncludedIndex <= rf.LastIncludedIndex || args.LastIncludedIndex < rf.commitIndex {
	if args.LastIncludedIndex <= rf.LastIncludedIndex {
		////	log.Printf("{machine %v} InstallSnapshot invalid", rf.me)
		return
	}

	new_logs := make([]Entry, 1)
	new_logs[0].Term = args.LastIncludedTerm

	// 快照后还有数据，需要保留
	if args.LastIncludedIndex < rf.GetlastLog().Index {
		new_logs = append(new_logs, rf.logs[args.LastIncludedIndex-rf.LastIncludedIndex+1:]...)
	}

	// 重置日志
	rf.LastIncludedIndex = args.LastIncludedIndex
	rf.LastIncludedTerm = args.LastIncludedTerm
	rf.logs = new_logs

	// 提交快照，给上层applier
	rf.persistStateAndSnapshot(args.Data)
	applymsg := ApplyMsg{
		SnapshotValid: true,
		Snapshot:      args.Data,
		SnapshotTerm:  args.LastIncludedTerm,
		SnapshotIndex: args.LastIncludedIndex}
	rf.mu.Unlock()
	rf.applyCh <- applymsg
	rf.mu.Lock()

	rf.lastApplied = rf.LastIncludedIndex
	rf.commitIndex = rf.LastIncludedIndex

	reply.Xindex = rf.GetRealLastLogIndex() + 1
	reply.Success = true
	////log.Printf("{machine %v} install snapshot after, lastIncludedIndex: %v, lastIncludedTerm: %v",
	////	rf.me, rf.LastIncludedIndex, rf.LastIncludedTerm)
}

func (rf *Raft) sendInstallSnapshot(server int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	return ok
}

func (rf *Raft) installSnapshotLeader(serverId int, args *InstallSnapshotArgs) {
	reply := InstallSnapshotReply{}
	ok := rf.sendInstallSnapshot(serverId, args, &reply)
	if !ok {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if !reply.Success {
		// server 2
		if reply.Term > rf.currentTerm {
			rf.UpdateState(STATE_FOLLOWER)
			rf.currentTerm = reply.Term
			rf.voteFor = -1
			rf.persistState()
			return
		}

		// 无效的快照
		rf.nextIndex[serverId] = reply.Xindex
		////log.Printf("{leader %v},失败提交 snapshot 到{machine %v}，nextIndex改为: %v", rf.me, serverId, rf.nextIndex[serverId])
	} else {
		rf.nextIndex[serverId] = reply.Xindex
		rf.matchIndex[serverId] = reply.Xindex - 1
		////log.Printf("{leader %v},成功提交 snapshot 到{machine %v}，nextIndex改为: %v", rf.me, serverId, rf.nextIndex[serverId])
	}

}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
// Call() false 可能是有一个服务器挂了
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.

// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.

// start agreement on a new log entry
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//term, isLeader = rf.GetState() 死锁
	if rf.state != STATE_LEADER {
		return index, term, false
	}
	////log.Printf("+++++++{leader %v} start", rf.me)
	index = rf.GetRealLastLogIndex() + 1
	term = rf.currentTerm
	newlog := Entry{
		Index: index,
		Term:  term,
		Cmd:   command,
	}
	rf.logs = append(rf.logs, newlog)
	rf.persistState()
	////log.Printf("{leader %v} append log, length %v!!!!!!!!!", rf.me, len(rf.logs))
	//log.Printf("{leader %v} append log: %v, length %v!!!!!!!!!", rf.me, newlog, len(rf.logs))
	// rf.broadcastLeader(false)

	return index, term, isLeader
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
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	////log.Printf("kill raft id: %v\n", rf.me)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for !rf.killed() {
		time.Sleep(rf.heartbeatTime)
		//log.Printf("{machine %v} Now time : %v ,Elelcion time : %v", rf.me, time.Now(), rf.electionTime)
		rf.mu.Lock()

		// Your code here (2A)
		// Check if a leader election should be started.
		// follower 2
		if time.Now().After(rf.electionTime) {
			rf.startElection()
		}
		if rf.state == STATE_LEADER {
			rf.broadcastLeader(true)
		}

		rf.mu.Unlock()

	}
}

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
// the service or tester wants to create a Raft server. the ports of all the Raft servers (including this one) are in peers[]. this server's port is peers[me]. all the servers' peers[] arrayshave the same order. persister is a place for this server to save its persistent state, and also initially holds the most recent saved state, if any. applyCh is a channel on which the tester or service expects Raft to send ApplyMsg messages. Make() must return quickly, so it should start goroutines for any long-running work
/*服务或测试人员想要创建一个Raft服务器。所有Raft服务器（包括这一个）的端口
都在peers[]中。此服务器的端口是peers[me]。所有服务器的peers[]阵列具有相
同的顺序。persister是该服务器保存其持久状态的地方，并且最初还保存最近保存的
状态（如果有的话）。applyCh是测试人员或服务期望Raft在其上发送ApplyMsg消息
的通道。Make（）必须快速返回，因此它应该为任何长时间运行的工作启动goroutines
*/
// create a new Raft server instance
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	////log.Printf("{machine %v} make", rf.me)
	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	rf.state = STATE_FOLLOWER
	rf.voteFor = -1
	//rf.lastApplied = 0
	//rf.commitIndex = 0
	rf.LastIncludedIndex = 0
	rf.LastIncludedTerm = 0
	rf.applyVCond = sync.NewCond(&rf.mu)
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.applyCh = applyCh
	rf.persister = persister
	rf.logs = append(rf.logs, Entry{Index: 0, Term: 0}) // log 0
	rf.heartbeatTime = time.Duration(100) * time.Millisecond
	rf.resetElectionTime()
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	rf.lastApplied = rf.LastIncludedIndex
	rf.commitIndex = rf.LastIncludedIndex

	// start ticker goroutine to start elections
	go rf.ticker()

	go rf.applier()

	return rf
}

func (rf *Raft) resetElectionTime() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))              // 获取纳秒级时间戳作为随机数生成器的种子
	lasttime := time.Millisecond * time.Duration((r.Intn(200) + 200)) // 随机时间是（200-500）毫秒
	now := time.Now()
	rf.electionTime = now.Add(lasttime)
	// log.Printf("**********{machine %v} befor timt: %v, Now time : %v ,Elelcion time : %v", rf.me, now, time.Now(), rf.electionTime)
}

// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidataId  int
	LastLogIndex int
	LastLogTerm  int
}

// RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool // 如果canditate当选为true
}

// 投票RPC的接收函数
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term, reply.VoteGranted = rf.currentTerm, false

	// RequestVote RPC 1
	//if args.Term < rf.currentTerm || (args.Term == rf.currentTerm && rf.voteFor != -1 && rf.voteFor != args.CandidataId) {
	if args.Term < rf.currentTerm {
		// reply.Term, reply.VoteGranted = rf.currentTerm, false
		return
	}
	// all machine 2
	if args.Term > rf.currentTerm {
		rf.UpdateState(STATE_FOLLOWER)
		rf.currentTerm = args.Term
		rf.voteFor = -1 // 每一轮选举只投一票, 设置在term增加的时候voteFor还原为-1
		rf.persistState()

		reply.Term = rf.currentTerm // 更新reply.Term
	}

	// 保证包含所有已提交的log才能成为leader
	lastLog := rf.GetlastLog()
	uptodate := args.LastLogTerm > lastLog.Term || (args.LastLogTerm == lastLog.Term && args.LastLogIndex >= rf.GetRealLastLogIndex())

	// RequestV ote RPC 2
	if (rf.voteFor == -1 || rf.voteFor == args.CandidataId) && uptodate {
		rf.resetElectionTime()
		reply.VoteGranted = true
		rf.voteFor = args.CandidataId
		rf.persistState()
	}

}

// 投票RPC的发送和处理回复函数
func (rf *Raft) requsetVotesCan(serverid int, args *RequestVoteArgs, voteCounter *int, becomeLeader *sync.Once) {

	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(serverid, args, &reply)
	////log.Printf("{machine %v} got vote from {machine %v},reply {%v}", args.CandidataId, serverid, reply.VoteGranted)
	if !ok || !reply.VoteGranted {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if reply.Term > rf.currentTerm {
		rf.UpdateState(STATE_FOLLOWER)
		rf.currentTerm = reply.Term
		rf.persistState()
		return
	}
	if reply.Term == rf.currentTerm {

		*voteCounter++
		if *voteCounter > len(rf.peers)/2 {
			// 发送心跳宣示
			becomeLeader.Do(func() {
				////log.Printf("*********{machine %v} gets majority votes,term %v", rf.me, rf.currentTerm)
				rf.UpdateState(STATE_LEADER)
				for id, _ := range rf.peers {
					// 初始化日志同步参数
					rf.nextIndex[id] = rf.GetRealLastLogIndex() + 1
					rf.matchIndex[id] = 0 // 这里应该是server.lastIncludedindex 但取不到
				}
				rf.broadcastLeader(true) // empty appendEntriesRPC
			})
		}
	}

}

// election超时 开始选举
func (rf *Raft) startElection() {
	rf.currentTerm += 1
	// currentTermCan := rf.currentTerm
	rf.UpdateState(STATE_CANDIDATE)
	rf.voteFor = rf.me
	rf.persistState()

	rf.resetElectionTime()

	voteCounter := 1
	////log.Printf("-------{machine %v} start election,term %v\n", rf.me, rf.currentTerm)

	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidataId:  rf.me,
		LastLogIndex: rf.GetRealLastLogIndex(),
		LastLogTerm:  rf.GetlastLog().Term,
	}
	var becomeLeader sync.Once
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		go rf.requsetVotesCan(i, &args, &voteCounter, &becomeLeader)
	}
}

// AppendEntries RPC
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term    int
	Success bool // 如果follower包含prevLogIndex和preLogItem的日志
	Xindex  int  // 告诉leader期望下次发过来的nextIndex
	Xterm   int
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	//log.Printf("{follower %v}AppendEntries start args.Entries: %v", rf.me, args.Entries)

	reply.Term, reply.Success = rf.currentTerm, false
	lastLog := rf.GetlastLog()
	reply.Xindex, reply.Xterm = rf.GetRealLastLogIndex(), lastLog.Term
	// append entries 1
	if args.Term < rf.currentTerm {
		//reply.Term, reply.Success = rf.currentTerm, false
		return
	}

	// all server 2
	if args.Term > rf.currentTerm {
		rf.UpdateState(STATE_FOLLOWER)
		rf.currentTerm = args.Term
		rf.persistState()
	}

	rf.resetElectionTime()

	// candidate 3
	if rf.state == STATE_CANDIDATE {
		rf.UpdateState(STATE_FOLLOWER)
	}

	// append entries 2
	real_lastlog_index := rf.GetRealLastLogIndex()

	////log.Printf("{machine %v} term %v AppendEntries,PrevLogIndex %v,real_lastlog_index: %v,LastIncludedIndex: %v", rf.me, rf.currentTerm, args.PrevLogIndex, real_lastlog_index, rf.LastIncludedIndex)

	if real_lastlog_index < args.PrevLogIndex {
		// 这时候期待下次发过来的nextIndex就是lastlog的后一个
		reply.Xindex, reply.Xterm = real_lastlog_index+1, -1
		////log.Printf("{machine %v} term %v ,期望取PrevLogIndex %v ,但是reallastIndex %v，Xindex: %v,Xterm: %v",
		////	rf.me, args.Term, args.PrevLogIndex, real_lastlog_index, reply.Xindex, reply.Xterm)
		return
	}
	xterm := rf.logs[args.PrevLogIndex-rf.LastIncludedIndex].Term
	//log.Printf("{machin %v} xterm %v", rf.me, xterm)
	if xterm != args.PrevLogTerm {
		// 优化 term第一个index返回
		firstIndex := args.PrevLogIndex
		for firstIndex-rf.LastIncludedIndex > 0 && rf.logs[firstIndex-rf.LastIncludedIndex].Term == xterm {
			firstIndex--
		}

		reply.Xterm, reply.Xindex = xterm, firstIndex+1

		////log.Printf("{machine %v} term %v ,期望取term %v ,但是到index %v 都是 term %v", rf.me, args.Term, args.PrevLogTerm, firstIndex, xterm)
		return
	}
	// 日志一致性更新
	////log.Printf("{machin %v} term %v ,loglen: %v----append before,",
	////	rf.me, rf.currentTerm, len(rf.logs))
	//log.Printf("{machin %v} term %v ,loglen: %v----append before,Entries: %v,logs: %v",
	//	rf.me, rf.currentTerm, len(rf.logs), args.Entries, rf.logs)

	// append entries 3没考虑，bug 可能会受到leader过时的rpc，导致误删了之后的log
	// if xterm == args.PrevLogTerm {
	// 	rf.logs = rf.logs[:args.PrevLogIndex+1]
	// 	rf.logs = append(rf.logs, args.Entries...)
	// }
	// reply.Xindex, reply.Xterm = lastLog.Index, lastLog.Term

	//if xterm == args.PrevLogTerm {
	for idx, entry := range args.Entries {
		// append entries 3
		if entry.Index <= lastLog.Index && rf.logs[entry.Index-rf.LastIncludedIndex].Term != entry.Term {
			rf.logs = rf.logs[:entry.Index-rf.LastIncludedIndex]
			rf.persistState()
		}

		// append entries 4
		if entry.Index > rf.GetlastLog().Index {
			rf.logs = append(rf.logs, args.Entries[idx:]...)
			rf.persistState()
			break
		}

	}
	real_lastlog_index = rf.GetRealLastLogIndex()
	reply.Xindex, reply.Xterm = real_lastlog_index+1, -1 //期待下一次发过来的nextIndex就是lastlogIndex的下一个了
	//}
	////log.Printf("{machin %v} term %v ,loglen: %v,----append after", rf.me, rf.currentTerm, len(rf.logs))
	//log.Printf("{machin %v} term %v ,loglen: %v,----append after，logs: %v", rf.me, rf.currentTerm, len(rf.logs), rf.logs)

	// append entries 5
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommit, rf.GetRealLastLogIndex())
		// 应用到本server的状态机上
		////log.Printf("{follower %v}从append更新commitIndex: %v，开始同步应用状态机", rf.me, rf.commitIndex)
		rf.applyVCond.Broadcast()
	}
	reply.Success = true
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) appendEntriesLeader(serverId int, args *AppendEntriesArgs) {
	//log.Printf("{follower %v}appendEntriesLeader start args: %v", serverId, args)
	reply := AppendEntriesReply{}
	ok := rf.sendAppendEntries(serverId, args, &reply)
	if !ok {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// leader 3
	if !reply.Success {
		// server 2
		if reply.Term > args.Term {
			rf.UpdateState(STATE_FOLLOWER)
			rf.currentTerm = reply.Term
			rf.voteFor = -1
			rf.persistState()
			return
		}
		// 日志不一致，更新PrevLogindex，重发
		// if reply.Xindex < args.PrevLogIndex {
		// 	log.Printf("{leader %v} to {machine %v} inconsistent, PreLogIndex: %v, Xindex: %v", args.LeaderId, serverId, args.PrevLogIndex, reply.Xindex)
		// 	args.PrevLogIndex = rf.logs[reply.Xindex-1].Index
		// 	args.PrevLogTerm = rf.logs[reply.Xindex-1].Term
		// 	args.Entries = rf.logs[reply.Xindex:]
		// 	log.Printf("{leader %v} retry to {machine %v} PrevLogIndex: %v,PrevLogTerm: %v", rf.me, serverId, args.PrevLogIndex, args.PrevLogTerm)
		// 	// bug死锁了，重发RPC需要解锁
		// 	{
		// 		rf.mu.Unlock()
		// 		rf.appendEntriesLeader(serverId, args)
		// 		rf.mu.Lock()
		// 	}
		// 	return
		// }
		if reply.Xindex <= args.PrevLogIndex {
			////log.Printf("{leader %v} to {machine %v} inconsistent, PreLogIndex: %v, Xindex: %v", args.LeaderId, serverId, args.PrevLogIndex, reply.Xindex)
			rf.nextIndex[serverId] = reply.Xindex //返回的Xindex是不匹配Term的第一个index
			return
		}
	} else {
		// 如果成功且添加了新的log，更新nextIndex matchIndex
		rf.nextIndex[serverId] = reply.Xindex
		rf.matchIndex[serverId] = reply.Xindex - 1
		////log.Printf("{leader %v},成功提交 logs 到{machine %v}，开始检查可否commit", rf.me, serverId)
		rf.commitLeader()
	}

}

// leader广播所有follower心跳，进行日志同步
func (rf *Raft) broadcastLeader(heatbeat bool) {
	// log.Printf("{machine %v} start broadcast", rf.me)

	for id, _ := range rf.peers {
		if id == rf.me {
			rf.resetElectionTime()
			continue
		}
		// 需要同步 或者 心跳
		if rf.GetRealLastLogIndex() >= rf.nextIndex[id] || heatbeat {

			nextindex := rf.nextIndex[id]

			//******快照********
			if nextindex <= rf.LastIncludedIndex {
				////log.Printf("{leadder %v} rf.LastIncludedIndex[%v] try to installSnapshot {machine %v}, nextindex %v", rf.me, rf.LastIncludedIndex, id, nextindex)

				snapshot_args := InstallSnapshotArgs{
					Term:              rf.currentTerm,
					LeaderId:          rf.me,
					LastIncludedIndex: rf.LastIncludedIndex,
					LastIncludedTerm:  rf.LastIncludedTerm,
					Data:              rf.persister.snapshot}

				go rf.installSnapshotLeader(id, &snapshot_args) //快照之后不接append RPC，交给下一次心跳
				continue
			}

			//******心跳********
			if nextindex < 1 {
				nextindex = 1 // 1
			} else if nextindex > rf.GetRealLastLogIndex()+1 {
				////log.Printf("{leader %v} to {machine %v} nextindex[%v] >  rf.GetRealLastLogIndex() +1[%v],rf.LastIncludedIndex[%v]", rf.me, id, nextindex, rf.GetRealLastLogIndex()+1, rf.LastIncludedIndex)
				nextindex = rf.GetRealLastLogIndex() + 1 // last log
			}

			////log.Printf("{leadder %v} rf.LastIncludedIndex[%v] try to append {machine %v}, nextindex %v", rf.me, rf.LastIncludedIndex, id, nextindex)

			nextpreLog := rf.logs[nextindex-1-rf.LastIncludedIndex] // last pre log

			append_args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: Max(nextpreLog.Index, rf.LastIncludedIndex), // TODO之后如果取到0 说明要传RPC
				PrevLogTerm:  nextpreLog.Term,
				//Entries:      rf.logs[nextindex-rf.LastIncludedIndex:], // 先分配再copy复制，否则可能因为编译器优化原因会随切片变化
				Entries:      make([]Entry, len(rf.logs[nextindex-rf.LastIncludedIndex:])),
				LeaderCommit: rf.commitIndex,
			}
			copy(append_args.Entries, rf.logs[nextindex-rf.LastIncludedIndex:])
			//log.Printf("{leadder %v} try append RPC {machine %v}, Entries %v", rf.me, id, append_args.Entries)
			go rf.appendEntriesLeader(id, &append_args)
			//runtime.Gosched()
		}
	}
	//log.Printf("{machine %v} end broadcast", rf.me)
}

// leader检查日志一致性是否满足，如果满足就更新commitIndex
func (rf *Raft) commitLeader() {
	if rf.state != STATE_LEADER {
		return
	}
	////log.Printf("{leader %v} count log apply number,commitIndex: %v,lastlogIndex：%v", rf.me, rf.commitIndex, rf.GetlastLog().Index)
	for idx := rf.commitIndex + 1; idx <= rf.GetlastLog().Index; idx++ {
		if rf.logs[idx-rf.LastIncludedIndex].Term != rf.currentTerm {
			////log.Printf("rf.logs[%v].Term: %v != rf.currentTerm: %v", idx, rf.logs[idx-rf.LastIncludedIndex].Term, rf.currentTerm)
			continue
		}
		count := 1
		for ServerId, _ := range rf.peers {
			if ServerId != rf.me && rf.matchIndex[ServerId] >= idx {
				////log.Printf("rf.matchIndex[%v] : %v, try commit: %v", ServerId, rf.matchIndex[ServerId], idx)
				count++
				if count > len(rf.peers)/2 {
					rf.commitIndex = idx
					////log.Printf("{leader %v} add commitIndex to: %v, awake applyVCond", rf.me, idx)
					// rf.broadcastLeader(true)
					// time.Sleep(time.Duration(5) * time.Millisecond)
					rf.applyVCond.Broadcast()
					break
				}
			}
		}
		//log.Printf("count : %v", count)
	}

}

// server提交日志，应用操作到状态机
func (rf *Raft) applier() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	for !rf.killed() {

		// 继续 查是否需要commit
		////log.Printf("{machine %v} applyVCond awake,commitIndex: %v,lastApplied: %v", rf.me, rf.commitIndex, rf.lastApplied)
		if rf.commitIndex > rf.lastApplied && rf.commitIndex <= rf.GetlastLog().Index {
			rf.lastApplied++
			////log.Printf("{machine %v} commitIndex: %v,lastApplied: %v", rf.me, rf.commitIndex, rf.lastApplied)
			applymsg := ApplyMsg{
				CommandValid: true,
				CommandIndex: rf.lastApplied,
				Command:      rf.logs[rf.lastApplied-rf.LastIncludedIndex].Cmd}
			// chanl读写的时候不要上锁，因为可能有阻塞
			rf.mu.Unlock()
			rf.applyCh <- applymsg
			rf.mu.Lock()
			//log.Printf("~~~~~~~{machine %v} apply applymsgIndex: %v,applymsg: %v", rf.me, applymsg.CommandIndex, applymsg)
			////log.Printf("~~~~~~~{machine %v} apply applymsgIndex: %v", rf.me, applymsg.CommandIndex)
		} else {
			rf.applyVCond.Wait()
		}
	}

}
