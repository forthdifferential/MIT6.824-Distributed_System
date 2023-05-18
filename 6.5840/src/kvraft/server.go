package kvraft

import (
	"bytes"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"6.5840/labgob"
	"6.5840/labrpc"
	"6.5840/raft"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Optype   string //  Get / Put / Append
	Key      string
	Value    string
	ClientID int64
	RpcID    int
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate      int // snapshot if log grows this big
	lastIncludedIndex int // 快照的最后一个index
	// Your definitions here.
	database   map[string]string
	requsetTab map[int64]requestEntry
	//receiveVCond *sync.Cond
	chanMap map[int]chan string
}
type requestEntry struct {
	RpcID int
	Value string
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.

	// 检查leader
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Success, reply.Err = false, "not leader"
		return
	}

	// 重复检测
	if ok, value := kv.checkRpc(args.ClientID, args.RpcID); ok {
		reply.Success, reply.Value = true, value
		return
	}

	// start 日志添加
	timer := time.NewTimer(500 * time.Millisecond)

	option := Op{
		Optype:   "Get",
		Key:      args.Key,
		ClientID: args.ClientID,
		RpcID:    args.RpcID,
	}

	index, _, isLeader := kv.rf.Start(option)
	if !isLeader {
		reply.Success, reply.Err = false, "not leader"
		return
	}

	// 等待apply后回复

	ch := kv.getOrMakeCh(index)

	DPrintf("{S%v}等待一致性通过或超时,Optype: Get,{C%v},RpcID: %v,Key %v", kv.me, args.ClientID, args.RpcID, args.Key)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时, {C%v} rpcId: %v", kv.me, args.ClientID, args.RpcID)
		reply.Success, reply.Err = false, "time out"

	case val := <-ch:
		DPrintf("{S%v} 成功, {C%v} rpcId: %v", kv.me, args.ClientID, args.RpcID)
		reply.Success, reply.Value = true, val
	}

	kv.deleteCh(index)

}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.

	// 检查leader
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Success, reply.Err = false, "not leader"
		return
	}

	// 重复检测
	if ok, _ := kv.checkRpc(args.ClientID, args.RpcID); ok {
		DPrintf("重复RPC: %v", args.RpcID)
		reply.Success = true
		return
	}

	// start 日志添加
	timer := time.NewTimer(500 * time.Millisecond)
	option := Op{
		Optype:   args.Optype,
		Key:      args.Key,
		Value:    args.Value,
		ClientID: args.ClientID,
		RpcID:    args.RpcID,
	}

	index, _, isLeader := kv.rf.Start(option)
	if !isLeader {
		reply.Success, reply.Err = false, "not leader"
		return
	}

	// 等待apply后回复
	ch := kv.getOrMakeCh(index)
	DPrintf("{S%v}提交，等待一致性通过或超时,Optype: Get,{C%v},RpcID: %v,Key %v", kv.me, args.ClientID, args.RpcID, args.Key)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时, {C%v} rpcId: %v", kv.me, args.ClientID, args.RpcID)
		reply.Success, reply.Err = false, "time out"
	case <-ch:
		DPrintf("{S%v} 成功, {C%v} rpcId: %v", kv.me, args.ClientID, args.RpcID)
		reply.Success = true
	}
	timer.Stop()
	kv.deleteCh(index)
}

// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
// 当不再需要KVServer实例时，测试人员会调用Kill（）。
// 为了您的方便，我们提供了设置rf.dead的代码（不需要锁），以及在长时间运行的循环
// 中测试rf.dead的killed（）方法。您也可以将自己的代码添加到Kill（）中。
// 您不需要对此做任何操作，但可以方便地（例如）抑制Kill（）ed实例的调试输出。
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
	DPrintf("{server %v} kill", kv.me)
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
// servers[]包含一组服务器的端口，这些服务器将通过Raft进行协作以
// 形成容错K/V服务。me是服务器[]中当前服务器的索引。k/v服务器
// 应通过底层Raft实现存储快照，该实现应调用persister.SaveStateAndSnapshot（）
// 以原子方式将Raft状态与快照一起保存。k/v服务器应该在Raft的保存状态超过
// maxraftstate字节时进行快照，以便允许Raft垃圾收集其日志。如果maxraftstate为-1，
// 则不需要进行快照。StartKVServer（）必须快速返回，因此它应该启动goroutines
// 对于任何长时间运行的工作。
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.lastIncludedIndex = 0
	// You may need initialization code here.
	kv.database = make(map[string]string)
	kv.requsetTab = make(map[int64]requestEntry)
	kv.chanMap = make(map[int]chan string)
	kv.applyCh = make(chan raft.ApplyMsg, 1)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	// You may need initialization code here.

	kv.readSnapshot(kv.rf.GetSnapshot())

	go kv.receiver() // 接收raft apply的条目 并处理

	return kv
}

func (kv *KVServer) receiver() {
	// 阻塞的读取chanl
	for !kv.killed() {
		select {
		case recv_msg := <-kv.applyCh:
			// Op
			if recv_msg.CommandValid {
				DPrintf("//// {S%v} receive Op,index[%v]", kv.me, recv_msg.CommandIndex)
				if recv_msg.Command == nil {
					DPrintf("{S%v} 空的recv_msg[%v]", kv.me, recv_msg)
				}
				recvOp := recv_msg.Command.(Op)

				DPrintf("{S%v} receive Op,client[%v],RpcID:[%v],Key [%v]", kv.me, recvOp.ClientID, recvOp.RpcID, recvOp.Key)
				if recvOp.Optype != "Get" {
					DPrintf("看看重复检查有没有问题{S%v} receive Op,Optype[%v],Key [%v],Value[%v]", kv.me, recvOp.Optype, recvOp.Key, recvOp.Value)

				}
				// 检查快照
				DPrintf("{S%v}检查缓存大小", kv.me)
				if kv.maxraftstate != -1 && kv.rf.GetStateSize() >= kv.maxraftstate-8 {
					// 缓存接近上限，启动快照
					DPrintf("{S%v}缓存接近上限，快照", kv.me)
					if ok, snapshot := kv.generateSnapshot(recv_msg.CommandIndex - 1); ok { //存的是这个index应用之前的状态
						DPrintf("{S%v}快照完成，通知raft persist", kv.me)
						kv.rf.Snapshot(recv_msg.CommandIndex-1, snapshot)
					}

				}

				// 应用
				commentRet := kv.apply(recvOp)
				// DPrintf("SERVER {server %v} success apply Optype: %v,ClientID: %v,RpcID: %v,Key %v", kv.me, recvOp.Optype, recvOp.ClientID, recvOp.RpcID, recvOp.Key)

				// 唤醒 返回结果
				// 只有leader在处理RPC,只有通道还没被删除的时候需要返回
				if _, isleader := kv.rf.GetState(); isleader {

					if ch, ok := kv.getChIfHas(recv_msg.CommandIndex); ok {
						ch <- commentRet
					} else {
						DPrintf("{S%v} 没有这个通道了 {C%v},RpcID: %v", kv.me, recvOp.ClientID, recvOp.RpcID)
					}

				}

			}

			// Snapshot
			if recv_msg.SnapshotValid {
				DPrintf("//// {S%v} receive Snapshot,index: %v", kv.me, recv_msg.SnapshotIndex)
				if ok := kv.snapshotValid(recv_msg.SnapshotIndex); ok {
					DPrintf("{S%v}快照有效，开始应用到状态机", kv.me)
					kv.readSnapshot(recv_msg.Snapshot)
				}

			}
		}

	}

}

func (kv *KVServer) apply(recvOp Op) string {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	commentRet := "" // 操作结果

	// 检查是否重复执行
	if request_entry, ok := kv.requsetTab[recvOp.ClientID]; ok && recvOp.RpcID == request_entry.RpcID {
		// 已有的条目 直接返回结果
		DPrintf("!!!! {S%v} cmd already apply Optype: %v,{C%v},RpcID: %v,Key %v", kv.me, recvOp.Optype, recvOp.ClientID, recvOp.RpcID, recvOp.Key)
		if recvOp.Optype == "Get" {
			commentRet = request_entry.Value
		}
		return commentRet
	}

	// 新的条目
	// 执行
	if recvOp.Optype == "Put" {
		kv.database[recvOp.Key] = recvOp.Value
	} else if recvOp.Optype == "Append" {
		kv.database[recvOp.Key] += recvOp.Value
	}
	// 更新数据库
	newEntry := requestEntry{
		RpcID: recvOp.RpcID,
	}
	if recvOp.Optype == "Get" {
		if value, ok := kv.database[recvOp.Key]; ok {
			newEntry.Value = value
		} else {
			newEntry.Value = ""
		}
		commentRet = newEntry.Value
	}
	// 更新表
	kv.requsetTab[recvOp.ClientID] = newEntry
	DPrintf("{S%v}更新表kv.requsetTab[%v] latest RPCid: %v ", kv.me, recvOp.ClientID, newEntry.RpcID)
	return commentRet
}

// 重复检验，查找client对应的rpc是否已经执行
func (kv *KVServer) checkRpc(clientId int64, rpcId int) (bool, string) {
	// 检查表
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 已经执行RPC请求
	equset_entry, ok := kv.requsetTab[clientId]
	if ok && equset_entry.RpcID == rpcId {
		return true, equset_entry.Value
	}
	//DPrintf("重复检测，ok: %v, equset_entry.rpcID: %v == rpcId: %v ", ok, equset_entry.rpcID, rpcId)
	return false, ""
}

func (kv *KVServer) getOrMakeCh(index int) chan string {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, ok := kv.chanMap[index]; !ok {
		kv.chanMap[index] = make(chan string, 1)
		DPrintf("{S%v}创建chan[index: %v]", kv.me, index)
	}

	return kv.chanMap[index]

}

func (kv *KVServer) getChIfHas(index int) (chan string, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	ch, ok := kv.chanMap[index]
	if !ok {
		DPrintf("{S%v}不存在chan[index: %v]", kv.me, index)
	}
	return ch, ok
}

func (kv *KVServer) deleteCh(index int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.chanMap, index)

}

func (kv *KVServer) generateSnapshot(index int) (bool, []byte) {

	kv.mu.Lock()
	defer kv.mu.Unlock()

	if index <= kv.lastIncludedIndex {
		return false, nil
	}

	// 更新lastIncludedIndex
	kv.lastIncludedIndex = index

	newBuf := new(bytes.Buffer)
	newEncoder := labgob.NewEncoder(newBuf)

	err := newEncoder.Encode(kv.lastIncludedIndex)
	if err != nil {
		DPrintf("{S[%d]} encode lastIncludedIndex error: %v\n", kv.me, err)
		return false, nil
	}
	err = newEncoder.Encode(kv.requsetTab)
	if err != nil {
		DPrintf("{S[%d]} encode requsetTab error: %v\n", kv.me, err)
		return false, nil
	}
	err = newEncoder.Encode(kv.database)
	if err != nil {
		DPrintf("{S[%d]} encode database error: %v\n", kv.me, err)
		return false, nil
	}

	snapshot := newBuf.Bytes()

	DPrintf("{S%v} persist 快照,lastIncludedIndex[%v]", kv.me, kv.lastIncludedIndex)

	return true, snapshot
}

func (kv *KVServer) readSnapshot(data []byte) {

	if data == nil || len(data) < 1 { // bootstrap without any state? 空指针和空值
		return
	}

	newBuf := bytes.NewBuffer(data)
	newDecoder := labgob.NewDecoder(newBuf)
	var old_lastincludedindex int
	var old_requsetTab map[int64]requestEntry
	var old_database map[string]string

	kv.mu.Lock()
	defer kv.mu.Unlock()

	if newDecoder.Decode(&old_lastincludedindex) != nil ||
		newDecoder.Decode(&old_requsetTab) != nil ||
		newDecoder.Decode(&old_database) != nil {
		// 无法解码
		DPrintf("{S%d}: Decode error", kv.me)
	} else {
		kv.lastIncludedIndex = old_lastincludedindex
		kv.requsetTab = old_requsetTab
		kv.database = old_database
		DPrintf("{S%d}: Decode success,old_lastincludedindex[%v]", kv.me, kv.lastIncludedIndex)
	}

}

func (kv *KVServer) snapshotValid(index int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	return index > kv.lastIncludedIndex
}
