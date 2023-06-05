package shardkv

import (
	"bytes"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"6.5840/labgob"
	"6.5840/labrpc"
	"6.5840/raft"
	"6.5840/shardctrler"
)

const Debug = true

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
	// TODO 配置修改后切片转移
	Key    string
	Value  string
	Optype string // Put Append Get ShardOut ShardIn

	ClientID int64
	RpcID    int

	ConfigNum        int
	Config           shardctrler.Config
	Shards           []int
	ShardsData       map[int](map[string]string)
	ShardsRequestTab map[int](map[int64]requestEntry)
}

// 返回给RPC的chan传输数据类型
type ChanReply struct {
	Err   string
	Value string
}

type ShardKV struct {
	mu           sync.Mutex
	me           int
	rf           *raft.Raft
	applyCh      chan raft.ApplyMsg
	make_end     func(string) *labrpc.ClientEnd
	gid          int
	ctrlers      []*labrpc.ClientEnd
	maxraftstate int // snapshot if log grows this big
	dead         int32
	// Your definitions here.
	mck               *shardctrler.Clerk                          // client talk to the shardctrler
	lastIncludedIndex int                                         // 快照的最后一个index
	database          [shardctrler.NShards]map[string]string      // 分片的数据
	requestTab        [shardctrler.NShards]map[int64]requestEntry // 分片的请求表
	chanMap           map[int]chan ChanReply

	curConfig shardctrler.Config // 当前配置
	oldConfig shardctrler.Config // 上一个配置状态

	//outStates      [shardctrler.NShards]bool //shard迁出状态
	//inStates       [shardctrler.NShards]bool //shard迁入状态
	shardStates [shardctrler.NShards]string // 每个shard的状态
	// configchanging int32
}

// 客户请求重复检验记录条目类型
type requestEntry struct {
	RpcID int
	Value string
}

func (kv *ShardKV) Get(args *GetArgs, reply *GetReply) {
	// Your code here.

	// 检查配置，分片是否为当前group管理,且没有正在迁入迁出操作
	shard := key2shard(args.Key)

	if !kv.checkShard(shard) {
		reply.Err = ErrWrongGroup
		DPrintf("[Gid:%v]{S%v} Get, Client[%v],Rpc[%v],分片还没准备好", kv.gid, kv.me, args.ClientId, args.RpcId)
		return
	}

	// 检查leader
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 重复检测
	if ok, value := kv.checkRpc(args.ClientId, args.RpcId, shard); ok {
		reply.Err, reply.Value = OK, value
		return
	}

	// start 日志添加
	timer := time.NewTimer(GetOutTime)

	option := Op{
		Optype:   "Get",
		Key:      args.Key,
		ClientID: args.ClientId,
		RpcID:    args.RpcId,
	}

	index, _, isLeader := kv.rf.Start(option)
	if !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 等待apply后回复

	ch := kv.getOrMakeCh(index)

	DPrintf("->///[Gid:%v]{S%v}等待一致性通过或超时,Optype: Get,{C%v},RpcID: %v,Key %v", kv.gid, kv.me, args.ClientId, args.RpcId, args.Key)

	select {
	case <-timer.C:
		DPrintf("[Gid:%v]{S%v} Get超时, {C%v} rpcId: %v", kv.gid, kv.me, args.ClientId, args.RpcId)
		reply.Err = ErrTimeOut

	case chanReply := <-ch:
		if chanReply.Err == ErrWrongGroup {
			reply.Err = ErrWrongGroup
		} else if chanReply.Err == ErrNoKey {
			reply.Err = ErrNoKey
		} else {
			DPrintf("[Gid:%v]{S%v} Get成功, {C%v} rpcId: %v, key: %v, value: %v", kv.gid, kv.me, args.ClientId, args.RpcId, args.Key, chanReply.Value)
			reply.Err, reply.Value = OK, chanReply.Value
		}

	}

	kv.deleteCh(index)

}

func (kv *ShardKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.

	// 检查配置，shard是否为当前group管理,且没有正在迁入迁出操作
	shard := key2shard(args.Key)

	if !kv.checkShard(shard) {
		reply.Err = ErrWrongGroup
		DPrintf("[Gid:%v]{S%v} PutAppend失败ErrWrongGroup, checkShard{C%v} rpcId: %v", kv.gid, kv.me, args.ClientId, args.RpcId)
		return
	}

	// 检查leader
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 重复检测
	if ok, _ := kv.checkRpc(args.ClientId, args.RpcId, shard); ok {
		DPrintf("重复RPC: %v", args.RpcId)
		reply.Err = OK
		return
	}

	// start 日志添加
	timer := time.NewTimer(PutAppendOutTime)
	option := Op{
		Optype:   args.Optype,
		Key:      args.Key,
		Value:    args.Value,
		ClientID: args.ClientId,
		RpcID:    args.RpcId,
	}

	index, _, isLeader := kv.rf.Start(option)
	if !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 等待apply后回复
	ch := kv.getOrMakeCh(index)
	DPrintf("->///[Gid:%v]{S%v}提交，等待一致性通过或超时,Optype: PutAppend,{C%v},RpcID: %v,Key %v", kv.gid, kv.me, args.ClientId, args.RpcId, args.Key)

	select {
	case <-timer.C:
		DPrintf("[Gid:%v]{S%v} PutAppend超时, {C%v} rpcId: %v", kv.gid, kv.me, args.ClientId, args.RpcId)
		reply.Err = ErrTimeOut
	case chanReply := <-ch:
		if chanReply.Err == ErrWrongGroup {
			DPrintf("[Gid:%v]{S%v} PutAppend失败ErrWrongGroup, chanReply.Err{C%v} rpcId: %v", kv.gid, kv.me, args.ClientId, args.RpcId)
			reply.Err = ErrWrongGroup
		} else {
			DPrintf("[Gid:%v]{S%v} PutAppend成功, {C%v} rpcId: %v", kv.gid, kv.me, args.ClientId, args.RpcId)
			reply.Err = OK
		}
	}

	timer.Stop()
	kv.deleteCh(index)

}

func (kv *ShardKV) applier() {
	// 阻塞的读取chanl
	for !kv.killed() {
		select {
		case recv_msg := <-kv.applyCh:
			// Op
			if recv_msg.CommandValid {
				recvOp := recv_msg.Command.(Op)
				DPrintf("<-//// [Gid:%v]{S%v} receive Op,Optype:[%v], index[%v]", kv.gid, kv.me, recvOp.Optype, recv_msg.CommandIndex)

				// 应用
				// commentRet := kv.apply(recvOp)
				//DPrintf("[Gid:%v]{S%v}开始应用", kv.gid, kv.me)
				var commentRet ChanReply
				switch recvOp.Optype {
				case "Put", "Append", "Get":
					DPrintf("[Gid:%v]{S%v} ClientRequest", kv.gid, kv.me)
					commentRet = kv.applyClientRequest(recvOp)

					// 唤醒 返回结果 只有leader在处理RPC,只有通道还没被删除的时候需要返回
					if _, isleader := kv.rf.GetState(); isleader {
						if ch, ok := kv.getChIfHas(recv_msg.CommandIndex); ok {
							DPrintf("[Gid:%v]{S%v}ch <- commentRet，ClientRequest开始添加111", kv.gid, kv.me)
							ch <- commentRet
							DPrintf("[Gid:%v]{S%v}ch <- commentRet，ClientRequest添加完成222", kv.gid, kv.me)
						} else {
							DPrintf("[Gid:%v]{S%v} 没有这个通道了 {C%v},RpcID: %v", kv.gid, kv.me, recvOp.ClientID, recvOp.RpcID)
						}
					}
				case "ConfigChange":
					DPrintf("[Gid:%v]{S%v} case ConfigChange", kv.gid, kv.me)
					kv.applyConfigChange(recvOp)
				case "AppendShards":
					DPrintf("[Gid:%v]{S%v} case AppendShards", kv.gid, kv.me)
					kv.applyAppendShards(recvOp)
				case "NoWaitShards":
					DPrintf("[Gid:%v]{S%v} case NoWaitShards", kv.gid, kv.me)
					kv.applyNoWaitShards(recvOp)
				case "DelShards":
					DPrintf("[Gid:%v]{S%v} case DelShards", kv.gid, kv.me)
					commentRet = kv.applyDelShards(recvOp)
					// 唤醒 返回结果 只有leader在处理RPC,只有通道还没被删除的时候需要返回
					if _, isleader := kv.rf.GetState(); isleader {
						if ch, ok := kv.getChIfHas(recv_msg.CommandIndex); ok {
							DPrintf("[Gid:%v]{S%v}ch <- commentRet，DelShardst 开始添加111", kv.gid, kv.me)
							ch <- commentRet
							DPrintf("[Gid:%v]{S%v}ch <- commentRet，DelShards 添加完成222", kv.gid, kv.me)
						} else {
							DPrintf("[Gid:%v]{S%v} 没有这个通道了 DelShards", kv.gid, kv.me)
						}
					}
				case "Empty":
					DPrintf("[Gid:%v]{S%v} case Empty", kv.gid, kv.me)
				}

				// 检查快照
				DPrintf("[Gid:%v]{S%v}检查缓存大小", kv.gid, kv.me)
				if kv.maxraftstate != -1 && kv.rf.GetStateSize() >= kv.maxraftstate {
					// 缓存接近上限，启动快照
					DPrintf("[Gid:%v]{S%v}缓存接近上限StateSize[%v]，快照", kv.gid, kv.me, kv.rf.GetStateSize())
					if ok, snapshot := kv.generateSnapshot(recv_msg.CommandIndex); ok { //存的是这个index应用之前的状态
						DPrintf("[Gid:%v]{S%v}快照完成，通知raft persist", kv.gid, kv.me)
						kv.rf.Snapshot(recv_msg.CommandIndex, snapshot)
					}
				}
				DPrintf("[Gid:%v]{S%v}快照完成，继续处理recv_msg.CommandIndex[%v]", kv.gid, kv.me, recv_msg.CommandIndex)
			}
			// Snapshot
			if recv_msg.SnapshotValid {
				DPrintf("<-//// [Gid:%v]{S%v} receive Snapshot,index: %v", kv.gid, kv.me, recv_msg.SnapshotIndex)
				if ok := kv.snapshotValid(recv_msg.SnapshotIndex); ok {
					DPrintf("[Gid:%v]{S%v}快照有效，开始应用到状态机", kv.gid, kv.me)
					kv.readSnapshot(recv_msg.Snapshot)
				}
			}
		}
	}
}

func (kv *ShardKV) applyClientRequest(recvOp Op) ChanReply {

	chanReply := ChanReply{Err: OK} // 操作结果
	shard := key2shard(recvOp.Key)
	// DPrintf("[Gid:%v]{S%v}等锁来才开始处理client request", kv.gid, kv.me)
	kv.mu.Lock()
	defer kv.mu.Unlock()

	DPrintf("[Gid:%v]{S%v}开始处理client request", kv.gid, kv.me)
	// 检查分片管理
	if !(kv.curConfig.Shards[shard] == kv.gid && (kv.shardStates[shard] == StateOk || kv.shardStates[shard] == StateWait)) {
		chanReply.Err = ErrWrongGroup
		DPrintf("[Gid:%v]{S%v}分片不能接受处理client request，kv.curConfig.Shards[shard][%v] != kv.gid[%v] || kv.shardStates[shard][%v]",
			kv.gid, kv.me, kv.curConfig.Shards[shard], kv.gid, kv.shardStates[shard])
		return chanReply
	}

	// 检查是否重复执行
	if request_entry, ok := kv.requestTab[shard][recvOp.ClientID]; ok && recvOp.RpcID == request_entry.RpcID {
		// 已有的条目 直接返回结果
		DPrintf("!!!! [Gid:%v]{S%v} cmd already apply Optype: %v,{C%v},RpcID: %v,Key %v", kv.gid, kv.me, recvOp.Optype, recvOp.ClientID, recvOp.RpcID, recvOp.Key)
		if recvOp.Optype == "Get" {
			chanReply.Value = request_entry.Value
		}
		return chanReply
	}

	// 执行新的条目，更新数据库
	newEntry := requestEntry{
		RpcID: recvOp.RpcID,
	}

	switch recvOp.Optype {
	case "Put":
		// DPrintf("[Gid:%v]{S%v} kv.database[shard%v][recvOp.Key%v] = recvOp.Value%v", kv.gid, kv.me, shard, recvOp.Key, recvOp.Value)
		kv.database[shard][recvOp.Key] = recvOp.Value
	case "Append":
		kv.database[shard][recvOp.Key] += recvOp.Value
	case "Get":
		if value, ok := kv.database[shard][recvOp.Key]; ok {
			newEntry.Value = value
		} else {
			newEntry.Value = ""
			chanReply.Err = ErrNoKey
		}
		chanReply.Value = newEntry.Value
	}
	// 更新请求表
	kv.requestTab[shard][recvOp.ClientID] = newEntry
	DPrintf("[Gid:%v]{S%v}更新表kv.requestTab[%v] latest RPCid: %v ", kv.gid, kv.me, recvOp.ClientID, recvOp.RpcID)
	return chanReply
}

func (kv *ShardKV) applyConfigChange(recvOp Op) {
	// DPrintf("[Gid:%v]{S%v}等锁来才开始应用ConfigChange", kv.gid, kv.me)
	kv.mu.Lock()
	defer kv.mu.Unlock()

	DPrintf("[Gid:%v]{S%v}开始应用ConfigChange", kv.gid, kv.me)

	// 重复检查 检查是否能修改config
	if recvOp.Config.Num != kv.curConfig.Num+1 {
		DPrintf("[Gid:%v]{S%v} applyConfigChange失败recvOp.Config.Num[%v] != kv.curConfig.Num[%v]+1", kv.gid, kv.me, recvOp.Config.Num, kv.curConfig.Num)
		return
	}

	// 上一次修改配置的拉取和删除shard完成
	for i := 0; i < shardctrler.NShards; i++ {
		if kv.shardStates[i] != StateOk {
			DPrintf("[Gid:%v]{S%v} applyConfigChange失败shard[%v]还没拉取或删除完成", kv.gid, kv.me, i)
			return
		}
	}

	// 执行修改配置
	// 初次修改不用拉取，不需要修改inStates,直接接受配置就行
	if recvOp.Config.Num != 1 {
		for i := 0; i < shardctrler.NShards; i++ {
			if recvOp.Config.Shards[i] != kv.curConfig.Shards[i] {
				// 开始等待移除旧shard
				if kv.curConfig.Shards[i] == kv.gid {
					kv.shardStates[i] = StateOut
					DPrintf("[Gid:%v]{S%v} applyConfigChange中，shard[%v]改为StateOut", kv.gid, kv.me, i)
				}
				// 开始等待添加新shard
				if recvOp.Config.Shards[i] == kv.gid {
					kv.shardStates[i] = StateIn
					DPrintf("[Gid:%v]{S%v} applyConfigChange中，shard[%v]改为StateIn", kv.gid, kv.me, i)

					// 不采用异步的方式直接拉取，因为可能会server会在go程时宕机，就永远无法拉取，改为定时检测拉取shard
					// go kv.pullShardLeader(i)
				}
			}
		}
	}
	kv.oldConfig = kv.curConfig
	kv.curConfig = recvOp.Config
	// kv.configchanging = 0
	// DPrintf("[Gid:%v]{S%v} kv.configchanging = 0", kv.gid, kv.me)

	DPrintf("+++[Gid:%v]{S%v} applyConfigChange成功 [%v]", kv.gid, kv.me, recvOp.Config.Num)
}

func (kv *ShardKV) applyAppendShards(recvOp Op) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 重复检查 配置num相同才能append
	// if recvOp.ConfigNum != kv.curConfig.Num || kv.shardStates[recvOp.Shard] != StateIn {
	// 	DPrintf("[Gid:%v]{S%v} applyAppendShard失败,ConfigNum不匹配或者已经接受过了，recvOp.ConfigNum[%v] != kv.curConfig.Num[%v],kv.shardStates[%v][%v]", kv.gid, kv.me, recvOp.ConfigNum, kv.curConfig.Num, recvOp.Shard, kv.shardStates[recvOp.Shard])
	// 	return
	// }
	// 重复检查 配置num相同才能append
	if recvOp.ConfigNum != kv.curConfig.Num {
		DPrintf("[Gid:%v]{S%v} applyAppendShard失败,ConfigNum不匹配，recvOp.ConfigNum[%v] != kv.curConfig.Num[%v]", kv.gid, kv.me, recvOp.ConfigNum, kv.curConfig.Num)
		return
	}
	// 有的shard已经拿到了则不应用
	for _, shard := range recvOp.Shards {
		if kv.shardStates[shard] != StateIn {
			DPrintf("[Gid:%v]{S%v} applyAppendShard失败,curConfigNum[%v],kv.shardStates[shard%v][%v]", kv.gid, kv.me, kv.curConfig.Num, shard, kv.shardStates[shard])
			return
		}
	}

	// 拿到shards数据
	for _, shard := range recvOp.Shards {
		sharddata := make(map[string]string)
		for k, v := range recvOp.ShardsData[shard] {
			sharddata[k] = v
		}
		shardrequesttab := make(map[int64]requestEntry)
		for k, v := range recvOp.ShardsRequestTab[shard] {
			shardrequesttab[k] = v
		}

		kv.database[shard] = sharddata
		kv.requestTab[shard] = shardrequesttab
		kv.shardStates[shard] = StateWait // 应用完成，等待告知ex-ower删除这个shard对应的数据
	}

	// 需要raft层一致性通过

	//kv.inStates[recvOp.Shard] = false
	DPrintf("[Gid:%v]{S%v} applyAppendShard成功,recvOp.ConfigNum[%v],shards[%v]", kv.gid, kv.me, recvOp.ConfigNum, recvOp.Shards)

}

func (kv *ShardKV) applyNoWaitShards(recvOp Op) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 重复检查 配置num相同才能append
	if recvOp.ConfigNum != kv.curConfig.Num {
		DPrintf("[Gid:%v]{S%v} applyNoWaitShards失败,ConfigNum不匹配，recvOp.ConfigNum[%v] != kv.curConfig.Num[%v]", kv.gid, kv.me, recvOp.ConfigNum, kv.curConfig.Num)
		return
	}
	// 重复检查 有的shard已经不在wait了则不应用
	for _, shard := range recvOp.Shards {
		if kv.shardStates[shard] != StateWait {
			DPrintf("[Gid:%v]{S%v}  applyNoWaitShards失败,curConfigNum[%v],kv.shardStates[shard%v][%v]", kv.gid, kv.me, kv.curConfig.Num, shard, kv.shardStates[shard])
			return
		}
	}
	// 删除shard对应的数据,修改shard状态为OK
	for _, shard := range recvOp.Shards {
		kv.shardStates[shard] = StateOk
	}

	DPrintf("[Gid:%v]{S%v} applyNoWaitShards成功,recvOp.ConfigNum[%v],shards[%v]", kv.gid, kv.me, recvOp.ConfigNum, recvOp.Shards)
}

func (kv *ShardKV) applyDelShards(recvOp Op) ChanReply {
	chanReply := ChanReply{Err: OK}
	DPrintf("[Gid:%v]{S%v} applyDelShards等锁来", kv.gid, kv.me)
	kv.mu.Lock()
	defer kv.mu.Unlock()
	DPrintf("[Gid:%v]{S%v} applyDelShards开始", kv.gid, kv.me)
	// 重复检查 配置num相同才能append
	if recvOp.ConfigNum != kv.curConfig.Num {
		DPrintf("[Gid:%v]{S%v} applyDelShards失败,ConfigNum不匹配，recvOp.ConfigNum[%v] != kv.curConfig.Num[%v]", kv.gid, kv.me, recvOp.ConfigNum, kv.curConfig.Num)
		return chanReply
	}
	// 重复检查 有的shard已经删除了则不应用
	for _, shard := range recvOp.Shards {
		if kv.shardStates[shard] != StateOut {
			DPrintf("[Gid:%v]{S%v} applyAppendShard失败,curConfigNum[%v],kv.shardStates[shard%v][%v]", kv.gid, kv.me, kv.curConfig.Num, shard, kv.shardStates[shard])
			return chanReply
		}
	}
	// 删除shard对应的数据,修改shard状态为OK
	for _, shard := range recvOp.Shards {
		//kv.database[shard] = map[string]string{}
		//kv.requestTab[shard] = map[int64]requestEntry{}
		kv.database[shard] = nil
		kv.requestTab[shard] = nil
		kv.shardStates[shard] = StateOk
	}

	DPrintf("[Gid:%v]{S%v} applyDelShards成功,recvOp.ConfigNum[%v],shards[%v]", kv.gid, kv.me, recvOp.ConfigNum, recvOp.Shards)
	return chanReply
}

// the tester calls Kill() when a ShardKV instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
func (kv *ShardKV) Kill() {
	// Your code here, if desired.
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	DPrintf("[Gid:%v]{S%v} kill", kv.gid, kv.me)
}
func (kv *ShardKV) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

// servers[] contains the ports of the servers in this group.
//
// me is the index of the current server in servers[].
//
// the k/v server should store snapshots through the underlying下层 Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
//
// the k/v server should snapshot when Raft's saved state exceeds
// maxraftstate bytes, in order to allow Raft to garbage-collect its
// log. if maxraftstate is -1, you don't need to snapshot.
//
// gid is this group's GID, for interacting with the shardctrler.
//
// pass ctrlers[] to shardctrler.MakeClerk() so you can send
// RPCs to the shardctrler.
//
// make_end(servername) turns a server name from a
// Config.Groups[gid][i] into a labrpc.ClientEnd on which you can
// send RPCs. You'll need this to send RPCs to other groups.
//
// look at client.go for examples of how to use ctrlers[]
// and make_end() to send RPCs to the group owning a specific shard.
//
// StartServer() must return quickly, so it should start goroutines
// for any long-running work.
func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int, gid int, ctrlers []*labrpc.ClientEnd, make_end func(string) *labrpc.ClientEnd) *ShardKV {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})

	kv := new(ShardKV)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.make_end = make_end
	kv.gid = gid
	kv.ctrlers = ctrlers

	// Your initialization code here.
	kv.lastIncludedIndex = 0
	// map需要初始化
	kv.database = [shardctrler.NShards]map[string]string{}
	kv.requestTab = [shardctrler.NShards]map[int64]requestEntry{}
	kv.shardStates = [shardctrler.NShards]string{}

	for i := 0; i < shardctrler.NShards; i++ {
		kv.database[i] = make(map[string]string)
		kv.requestTab[i] = make(map[int64]requestEntry)
		kv.shardStates[i] = StateOk
	}
	kv.chanMap = make(map[int]chan ChanReply)
	// Use something like this to talk to the shardctrler:
	kv.mck = shardctrler.MakeClerk(kv.ctrlers)

	kv.applyCh = make(chan raft.ApplyMsg, 1)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	// kv.outStates = [shardctrler.NShards]bool{}
	// kv.inStates = [shardctrler.NShards]bool{}

	// kv.configchanging = 0
	// DPrintf("[Gid:%v]{S%v} 初始化kv.configchanging = 0", kv.gid, kv.me)
	kv.curConfig = shardctrler.Config{Num: 0, Groups: map[int][]string{}}
	DPrintf("[Gid:%v]{S%v} StartServer：maxraftstate[%v]", kv.gid, kv.me, maxraftstate)

	kv.readSnapshot(kv.rf.GetSnapshot())

	go kv.applier()

	go kv.configChangeChecker()
	go kv.shardsPullChecker()
	go kv.shardDelChecker()

	go kv.curTermLogEmptyChecker()

	return kv
}

// 定期检查是否可以应用下一次config
func (kv *ShardKV) configChangeChecker() {

	for !kv.killed() {
		time.Sleep(50 * time.Millisecond)
		// 只有leader发起配置更换
		if _, isLeader := kv.rf.GetState(); !isLeader {
			continue
		}
		DPrintf("[Gid:%v]{S%v} configChangeChecker 开始检查配置", kv.gid, kv.me)
		// 新的配置还在等待应用
		ok, curConfigNum := kv.configChangeable()
		if !ok {
			DPrintf("[Gid:%v]{S%v} configChangeChecker 不能更改新配置", kv.gid, kv.me)
			continue
		}

		DPrintf("[Gid:%v]{S%v}开始 主动拉取curConfig[%v] + 1", kv.gid, kv.me, curConfigNum)
		newConfig := kv.mck.Query(curConfigNum + 1)
		DPrintf("[Gid:%v]{S%v}完成 主动拉取curConfig[%v] + 1,newConfig[%v]", kv.gid, kv.me, curConfigNum, newConfig)
		// 配置已经是最新的

		if newConfig.Num <= curConfigNum {
			DPrintf("[Gid:%v]{S%v} configChangeChecker,已经是最新的 newConfig.Num[%v][完整config[%v]] <= curConfigNum[%v]", kv.gid, kv.me, newConfig.Num, newConfig, curConfigNum)
			continue
		}

		// 添加新的配置
		// atomic.AddInt32(&kv.configchanging, 1)
		// DPrintf("[Gid:%v]{S%v} kv.configchanging = 1", kv.gid, kv.me)

		option := Op{
			Optype: "ConfigChange",
			Config: newConfig,
		}
		kv.rf.Start(option)

		DPrintf("->///[Gid:%v]{S%v}等待一致性通过,Optype: configchange,newConfig:[%v]", kv.gid, kv.me, newConfig)
	}
}

// 定时检测是否需要拉取新的shards
func (kv *ShardKV) shardsPullChecker() {
	for !kv.killed() {
		time.Sleep(50 * time.Millisecond)
		// 只有leader发起配置更换
		if _, isLeader := kv.rf.GetState(); !isLeader {
			continue
		}
		gidtoShards, curConfigNum := kv.getPreGidtoShards(StateIn)

		DPrintf("[Gid:%v]{S%v} shardsPullChecker 开始,curConfigNum[%v],gidtoShards[%v]", kv.gid, kv.me, curConfigNum, gidtoShards)
		for gid, shards := range gidtoShards {
			// 按照需要拉取的group拉取shards
			go kv.pullShardsLeader(gid, shards, curConfigNum)
		}
	}
}

// 定时检测是否需要告知exower删除shards
func (kv *ShardKV) shardDelChecker() {
	for !kv.killed() {
		time.Sleep(50 * time.Millisecond)
		// 只有leader发起配置更换
		if _, isLeader := kv.rf.GetState(); !isLeader {
			continue
		}
		gidtoShards, curConfigNum := kv.getPreGidtoShards(StateWait)

		DPrintf("[Gid:%v]{S%v} shardDelChecker 开始,curConfigNum[%v],gidtoShards[%v]", kv.gid, kv.me, curConfigNum, gidtoShards)
		for gid, shards := range gidtoShards {
			go kv.delShardsLeader(gid, shards, curConfigNum)
		}

	}
}

// 定时检测 raft 层的 leader 是否拥有当前 term 的日志，如果没有则提交一条空日志，
// 这使得新 leader 的状态机能够迅速达到最新状态，从而避免多 raft 组间的活锁状态
func (kv *ShardKV) curTermLogEmptyChecker() {
	for !kv.killed() {
		time.Sleep(50 * time.Millisecond)
		// 只有leader发起配置更换
		if _, isLeader := kv.rf.GetState(); !isLeader {
			continue
		}
		if !kv.rf.HasLogInCurTerm() {
			option := Op{
				Optype: "Empty",
			}
			// 添加 shard 到 raft集群
			kv.rf.Start(option)
		}
	}
}

type DelShards_Args struct {
	Shards    []int
	ConfigNum int
}
type DelShards_Reply struct {
	Err string
}

func (kv *ShardKV) DelShards(args *DelShards_Args, reply *DelShards_Reply) {
	// 设置只能拉取leader的，来保证之后删除的条件一致性满足
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 重复检验，ConfigNum匹配，是否已经被删除
	kv.mu.Lock()
	DPrintf("[Gid:%v]{S%v} DelShards RPC 开始,shards[%v],configNum[%v]", kv.gid, kv.me, args.Shards, args.ConfigNum)

	// 已经删除并且进入下一个Config，过时的RPC，但不能保证RPC发送方不在wait，所以当做OK一样的效果
	if args.ConfigNum < kv.curConfig.Num {
		reply.Err = ErrConfigNumOut
		DPrintf("[Gid:%v]{S%v} DelShards RPC 过时,args.NewConfigNum[%v] != kv.curConfig.Num[%v],此时server信息：shardStates[%v]", kv.gid, kv.me, args.ConfigNum, kv.curConfig.Num, kv.shardStates)
		kv.mu.Unlock()
		return
	}
	// 更新的ConfigNum要求删除；
	// 如果RPC接受方处于 宕机恢复的过程中，
	if args.ConfigNum > kv.curConfig.Num {
		reply.Err = ErrConfigNumNotMatch
		DPrintf("[Gid:%v]{S%v} DelShards RPC ErrConfigNumNotMatch,args.NewConfigNum[%v] != kv.curConfig.Num[%v],此时server信息：shardStates[%v]", kv.gid, kv.me, args.ConfigNum, kv.curConfig.Num, kv.shardStates)
		kv.mu.Unlock()
		return
	}

	// 重复检测 如果已经删除了，返回删除完成，让发送方NoWait
	for _, shard := range args.Shards {
		if kv.shardStates[shard] != StateOut {
			reply.Err = OK
			DPrintf("[Gid:%v]{S%v} DelShards RPC 过时,此时server信息：shardStates[%v]", kv.gid, kv.me, kv.shardStates)
			kv.mu.Unlock()
			return
		}
	}
	kv.mu.Unlock()

	// 添加到raft层
	option := Op{
		Optype:    "DelShards",
		Shards:    args.Shards,
		ConfigNum: args.ConfigNum,
	}

	timer := time.NewTimer(GetOutTime)

	index, _, isLeader := kv.rf.Start(option)
	if !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	// 等待apply后回复
	ch := kv.getOrMakeCh(index)
	DPrintf("->///[Gid:%v]{S%v}等待一致性通过,Optype: DelShards，ConfigNum[%v],shards[%v]", kv.gid, kv.me, args.ConfigNum, args.Shards)

	select {
	case <-timer.C:
		DPrintf("[Gid:%v]{S%v}超时,Optype: DelShards，ConfigNum[%v],shards[%v]", kv.gid, kv.me, args.ConfigNum, args.Shards)
		reply.Err = ErrTimeOut

	case <-ch:
		DPrintf("[Gid:%v]{S%v}成功,Optype: DelShards，ConfigNum[%v],shards[%v]", kv.gid, kv.me, args.ConfigNum, args.Shards)
		reply.Err = OK
	}
	kv.deleteCh(index)
}

func (kv *ShardKV) delShardsLeader(gid int, shards []int, curConfigNum int) {

	// 只有leader来pull shards
	if _, isLeader := kv.rf.GetState(); !isLeader {
		return
	}
	args := DelShards_Args{Shards: shards, ConfigNum: curConfigNum}
	// 修改配置的时候，已经保证了需要拉取的一定是配置2开始，不用考虑没有curConfigNum-1的配置
	DPrintf("[Gid:%v]{S%v} 开始delShardsLeader Query，old_Config[%v-1]", kv.gid, kv.me, curConfigNum)
	// old_Config := kv.mck.Query(curConfigNum - 1)
	// 得到old_Config
	valid, old_Config := kv.getOldConfig(curConfigNum)
	if !valid {
		DPrintf("[Gid:%v]{S%v} delShardsLeader getOldConfig[%v],失效", kv.gid, kv.me, old_Config)
		return
	}
	DPrintf("[Gid:%v]{S%v} 完成 delShardsLeader Query，old_Config[%v]", kv.gid, kv.me, old_Config)

	for !kv.killed() {
		if servers, ok := old_Config.Groups[gid]; ok {
			// try each server for the shard.
			for si := 0; si < len(servers); si++ {
				srv := kv.make_end(servers[si])
				var reply DelShards_Reply
				DPrintf("[Gid:%v]{S%v} Call ShardKV.DelShard from[Gid%v]，NewConfigNum[%v]", kv.gid, kv.me, gid, curConfigNum)
				ok := srv.Call("ShardKV.DelShards", &args, &reply)
				// 考虑RPC失败的不确定性，就算回复是过时的RPC，也就是接收方已经删除了，发送方处理回复也需要添加一个Op到raft层
				// 把Op的有效性判断，交给apply
				if ok && (reply.Err == OK || reply.Err == ErrConfigNumOut) {
					// 修改删完的shard状态到raft中同步
					option := Op{
						Optype:    "NoWaitShards",
						Shards:    shards,
						ConfigNum: curConfigNum,
					}
					// 添加 shard 到 raft集群
					_, _, isLeader := kv.rf.Start(option)
					if !isLeader {
						DPrintf("no->///[Gid:%v]{S%v} delShardsLeader完成，但不再是leader", kv.gid, kv.me)
					} else {
						DPrintf("->///[Gid:%v]{S%v}等待一致性通过,Optype: NoWaitShards，shards[%v]", kv.gid, kv.me, shards)
					}
					return
				}
				if ok && (reply.Err == ErrConfigNumNotMatch) {
					// 找错组或者目标组已经给了
					DPrintf("[Gid:%v]{S%v} NoWaitShards ErrConfigNumNotMatch", kv.gid, kv.me)
					break
				}
			}
			// 只有leader来pull shards。可能这时候已经不是leader
			if _, isLeader := kv.rf.GetState(); !isLeader {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

	}
}

type PullShard_Args struct {
	Shards       []int
	NewConfigNum int
}
type PullShard_Reply struct {
	Err              Err
	NewConfigNum     int
	ShardsData       map[int](map[string]string)
	ShardsRequestTab map[int](map[int64]requestEntry)
}

// 拉取shard的RPC接受函数
func (kv *ShardKV) PullShards(args *PullShard_Args, reply *PullShard_Reply) {
	// 设置只能拉取leader的，来保证之后删除的条件一致性满足
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	kv.mu.Lock()
	defer kv.mu.Unlock()

	DPrintf("[Gid:%v]{S%v} 被PullShard 开始,shards[%v],configNum[%v]", kv.gid, kv.me, args.Shards, args.NewConfigNum)

	// RPC 接收方还没有更新到这个配置，让请求方再等等重新请求
	if args.NewConfigNum > kv.curConfig.Num {
		reply.Err = ErrConfigNumNotReady
		DPrintf("[Gid:%v]{S%v} 被PullShard ErrConfigNumNotReady,args.NewConfigNum[%v] != kv.curConfig.Num[%v],此时server信息：shardStates[%v]", kv.gid, kv.me, args.NewConfigNum, kv.curConfig.Num, kv.shardStates)
		return
	}

	// 过时的RPC 舍弃这个RPC
	// 因为被pull放只有 被通知可以删除整个shard后才会更新配置，所以被pull方配置更新则需求方已经拿到shard，说明只能是过时的RPC
	if args.NewConfigNum < kv.curConfig.Num {
		reply.Err = ErrConfigNumOut
		DPrintf("[Gid:%v]{S%v} 被PullShard ErrConfigNumOut,args.NewConfigNum[%v] != kv.curConfig.Num[%v],此时server信息：shardStates[%v]", kv.gid, kv.me, args.NewConfigNum, kv.curConfig.Num, kv.shardStates)
		return
	}

	reply.ShardsData = make(map[int]map[string]string)
	reply.ShardsRequestTab = make(map[int]map[int64]requestEntry)

	for _, shard := range args.Shards {
		sharddata := make(map[string]string)
		for k, v := range kv.database[shard] {
			sharddata[k] = v
		}
		shardrequesttab := make(map[int64]requestEntry)
		for k, v := range kv.requestTab[shard] {
			shardrequesttab[k] = v
		}
		reply.ShardsData[shard] = sharddata
		reply.ShardsRequestTab[shard] = shardrequesttab
	}

	reply.NewConfigNum = kv.curConfig.Num
	reply.Err = OK

	// 被拉取成功了，修改outStates
	// kv.outStates[args.Shard] = false

	DPrintf("[Gid:%v]{S%v} 被PullShard 完成,shards[%v],configNum[%v]", kv.gid, kv.me, args.Shards, args.NewConfigNum)
}

// 拉取shard的RPC发送函数
func (kv *ShardKV) pullShardsLeader(gid int, shards []int, curConfigNum int) {

	// 只有leader来pull shards
	if _, isLeader := kv.rf.GetState(); !isLeader {
		return
	}

	// 先拿到inState的shard，然后raft同步shard到组内
	args := PullShard_Args{Shards: shards, NewConfigNum: curConfigNum}

	// 修改配置的时候，已经保证了需要拉取的一定是配置2开始，不用考虑没有curConfigNum-1的配置
	DPrintf("[Gid:%v]{S%v} 开始 pullShardsLeader Query", kv.gid, kv.me)
	// old_Config := kv.mck.Query(curConfigNum - 1)
	valid, old_Config := kv.getOldConfig(curConfigNum)
	if !valid {
		DPrintf("[Gid:%v]{S%v} pullShardsLeader getOldConfig[%v],失效", kv.gid, kv.me, old_Config)
		return
	}
	DPrintf("[Gid:%v]{S%v} 完成 pullShardsLeader Query，old_Config[%v]", kv.gid, kv.me, old_Config)

	for !kv.killed() {
		if servers, ok := old_Config.Groups[gid]; ok {
			// try each server for the shard.
			for si := 0; si < len(servers); si++ {
				srv := kv.make_end(servers[si])
				var reply PullShard_Reply
				DPrintf("[Gid:%v]{S%v} Call ShardKV.PullShards from[Gid%v]，NewConfigNum[%v]", kv.gid, kv.me, gid, curConfigNum)
				ok := srv.Call("ShardKV.PullShards", &args, &reply)
				if ok && (reply.Err == OK) {
					// 添加得到的shard到raft中同步
					option := Op{
						Optype:           "AppendShards",
						ConfigNum:        reply.NewConfigNum,
						Shards:           shards,
						ShardsData:       reply.ShardsData,
						ShardsRequestTab: reply.ShardsRequestTab,
					}

					// 添加 shard 到 raft集群
					kv.rf.Start(option)
					DPrintf("->///[Gid:%v]{S%v}等待一致性通过,Optype: appendShards，shards[%v]", kv.gid, kv.me, shards)
					return
				} else if ok && (reply.Err == ErrConfigNumOut) {
					// 找错组或者目标组已经给了 接受方还没更新到这个config
					DPrintf("[Gid:%v]{S%v} pull shard ErrConfigNumOut", kv.gid, kv.me)
					return
				} else if ok && (reply.Err == ErrConfigNumNotReady) {
					DPrintf("[Gid:%v]{S%v} pull shard ErrConfigNumNotReady,等等在请求试试", kv.gid, kv.me)
					break
				}
				// ... not ok, or ErrWrongLeader
			}
			// 只有leader来pull shards。可能这时候已经不是leader
			if _, isLeader := kv.rf.GetState(); !isLeader {
				return
			}
			time.Sleep(100 * time.Millisecond)
			// ask controler for the latest configuration.
		}

	}
}

func (kv *ShardKV) generateSnapshot(index int) (bool, []byte) {
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
		DPrintf("[Gid:%v]{S[%d]} encode lastIncludedIndex error: %v\n", kv.gid, kv.me, err)
		return false, nil
	}
	err = newEncoder.Encode(kv.requestTab)
	if err != nil {
		DPrintf("[Gid:%v]{S[%d]} encode requestTab error: %v\n", kv.gid, kv.me, err)
		return false, nil
	}
	err = newEncoder.Encode(kv.database)
	if err != nil {
		DPrintf("[Gid:%v]{S[%d]} encode database error: %v\n", kv.gid, kv.me, err)
		return false, nil
	}

	err = newEncoder.Encode(kv.curConfig)
	if err != nil {
		DPrintf("[Gid:%v]{S[%d]} encode database error: %v\n", kv.gid, kv.me, err)
		return false, nil
	}
	// err = newEncoder.Encode(kv.oldConfig)
	// if err != nil {
	// 	DPrintf("[Gid:%v]{S[%d]} encode database error: %v\n", kv.gid, kv.me, err)
	// 	return false, nil
	// }

	err = newEncoder.Encode(kv.shardStates)
	if err != nil {
		DPrintf("[Gid:%v]{S[%d]} encode database error: %v\n", kv.gid, kv.me, err)
		return false, nil
	}

	snapshot := newBuf.Bytes()

	DPrintf("[Gid:%v]{S%v} persist 快照,lastIncludedIndex[%v]", kv.gid, kv.me, kv.lastIncludedIndex)
	return true, snapshot
}

func (kv *ShardKV) readSnapshot(data []byte) {

	if data == nil || len(data) < 1 { // bootstrap without any state? 空指针和空值
		return
	}

	newBuf := bytes.NewBuffer(data)
	newDecoder := labgob.NewDecoder(newBuf)
	var old_lastincludedindex int
	var old_requestTab [shardctrler.NShards]map[int64]requestEntry
	var old_database [shardctrler.NShards]map[string]string
	var old_curConfig shardctrler.Config
	//var old_oldConfig shardctrler.Config
	var old_shardStates [shardctrler.NShards]string
	//var old_configchanging int32

	kv.mu.Lock()
	defer kv.mu.Unlock()

	if newDecoder.Decode(&old_lastincludedindex) != nil ||
		newDecoder.Decode(&old_requestTab) != nil ||
		newDecoder.Decode(&old_database) != nil ||
		newDecoder.Decode(&old_curConfig) != nil ||
		//newDecoder.Decode(&old_oldConfig) != nil ||
		newDecoder.Decode(&old_shardStates) != nil {
		// 无法解码
		DPrintf("[Gid:%v]{S%d}: Decode error", kv.gid, kv.me)
	} else {
		kv.lastIncludedIndex = old_lastincludedindex
		kv.requestTab = old_requestTab
		kv.database = old_database
		kv.curConfig = old_curConfig
		//kv.oldConfig = old_oldConfig
		kv.shardStates = old_shardStates
		if old_curConfig.Num > 1 {
			kv.oldConfig = kv.mck.Query(old_curConfig.Num - 1)
		}
		DPrintf("[Gid:%v]{S%d}: 应用快照 success,old_lastincludedindex[%v]", kv.gid, kv.me, kv.lastIncludedIndex)
	}
}

func (kv *ShardKV) snapshotValid(index int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	return index > kv.lastIncludedIndex
}

// 重复检验，查找client对应的rpc是否已经执行
func (kv *ShardKV) checkRpc(clientId int64, rpcId int, shard int) (bool, string) {
	// 检查表
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 已经执行RPC请求
	equset_entry, ok := kv.requestTab[shard][clientId]
	if ok && equset_entry.RpcID == rpcId {
		return true, equset_entry.Value
	}
	return false, ""
}

// 检查shard是否可以接受客户端请求
func (kv *ShardKV) checkShard(shard int) bool {

	kv.mu.Lock()
	defer kv.mu.Unlock()
	DPrintf("[Gid:%v]{S%v} checkShard kv.curConfig.Shards[shard][%v] == kv.gid[%v] && kv.shardStates[shard][%v]", kv.gid, kv.me, kv.curConfig.Shards[shard], kv.gid, kv.shardStates[shard])
	return kv.curConfig.Shards[shard] == kv.gid && (kv.shardStates[shard] == StateOk || kv.shardStates[shard] == StateWait)
}

func (kv *ShardKV) getOrMakeCh(index int) chan ChanReply {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, ok := kv.chanMap[index]; !ok {
		kv.chanMap[index] = make(chan ChanReply, 1)
		DPrintf("[Gid:%v]{S%v}创建chan[index: %v]", kv.gid, kv.me, index)
	}

	return kv.chanMap[index]
}

func (kv *ShardKV) getChIfHas(index int) (chan ChanReply, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	ch, ok := kv.chanMap[index]
	if !ok {
		DPrintf("[Gid:%v]{S%v}不存在chan[index: %v]", kv.gid, kv.me, index)
	}
	return ch, ok
}

func (kv *ShardKV) deleteCh(index int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.chanMap, index)
}

// 得到对应state的shard在上一个config时所在的group
func (kv *ShardKV) getPreGidtoShards(specificState string) (map[int][]int, int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	gidtoShards := make(map[int][]int)
	curConfigNum := kv.curConfig.Num
	for shard, state := range kv.shardStates {
		if state == specificState {
			gid := kv.oldConfig.Shards[shard]
			gidtoShards[gid] = append(gidtoShards[gid], shard)
		}
	}
	return gidtoShards, curConfigNum
}

// 判断是否处于可修改配置的状态，没有处理数据迁入迁出
func (kv *ShardKV) configChangeable() (bool, int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	curConfigNum := kv.curConfig.Num

	for shard, state := range kv.shardStates {
		if state != StateOk {
			DPrintf("[Gid:%v]{S%v} configChangeable Err,shardStates[%v][%v]，configNum[%v]", kv.gid, kv.me, shard, state, kv.curConfig.Num)
			return false, curConfigNum
		}
	}
	return true, curConfigNum
}

func (kv *ShardKV) getOldConfig(curConfigNum int) (bool, shardctrler.Config) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	oldConfig := kv.oldConfig

	// 判断curConfigNum有没有过期
	if curConfigNum != kv.curConfig.Num {
		return false, oldConfig
	}

	return true, oldConfig
}

// func (kv *ShardKV) getCurConfig(curConfigNum int) (bool, shardctrler.Config) {
// 	kv.mu.Lock()
// 	defer kv.mu.Unlock()

// 	oldConfig := kv.oldConfig

// 	// 判断curConfigNum有没有过期
// 	if curConfigNum != kv.curConfig.Num {
// 		return false, oldConfig
// 	}

// 	return true, oldConfig
// }

// func (kv *ShardKV) appendShardLeader(reply PullShard_Reply, shards []int) {
// 	// 只有leader发起shard迁移
// 	if _, isLeader := kv.rf.GetState(); !isLeader {
// 		return
// 	}
// 	// RPC给的时候给副本，raft同步后server拿的时候拿副本，这里不需要用副本
// 	// // 用副本传输，防止race
// 	// sharddata := make(map[string]string)
// 	// for k, v := range reply.ShardData {
// 	// 	sharddata[k] = v
// 	// }
// 	// shardrequesttab := make(map[int64]requestEntry)
// 	// for k, v := range reply.ShardRequestTab {
// 	// 	shardrequesttab[k] = v
// 	// }
// 	option := Op{
// 		Optype:           "AppendShard",
// 		ConfigNum:        reply.NewConfigNum,
// 		Shards:           shards,
// 		ShardsData:       reply.ShardsData,
// 		ShardsRequestTab: reply.ShardsRequestTab,
// 	}

// 	// 添加 shard 到 raft集群
// 	kv.rf.Start(option)
// 	DPrintf("->///[Gid:%v]{S%v}等待一致性通过,Optype: appendShard，shards[%v]", kv.gid, kv.me, shards)
// }

// func (kv *ShardKV) ShardOut(shardNum int) {

// 	// start 日志添加
// 	timer := time.NewTimer(GetOutTime)
// 	option := Op{
// 		Optype: "ShardOut",
// 		Shard:  shardNum,
// 	}

// 	index, _, isLeader := kv.rf.Start(option)
// 	if !isLeader {
// 		return
// 	}

// 	// 等待apply后回复

// 	ch := kv.getOrMakeCh(index)

// 	DPrintf("[Gid:%v]{S%v}等待一致性通过或超时,Optype: ShardOut,Shard: %v", kv.gid, kv.me, shardNum)

// 	select {
// 	case <-timer.C:
// 		DPrintf("[Gid:%v]{S%v} ShardOut超时,Shard: %v", kv.gid, kv.me, shardNum)
// 	case <-ch:
// 		DPrintf("[Gid:%v]{S%v} ShardOut成功, Shard: %v", kv.gid, kv.me, shardNum)
// 	}

// 	kv.deleteCh(index)

// 	// 得到下一个config后发起配置改变
// }
