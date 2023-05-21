package shardctrler

import (
	"log"
	"sort"
	"sync"
	"time"

	"6.5840/labgob"
	"6.5840/labrpc"
	"6.5840/raft"
)

const Debug = true

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type ShardCtrler struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	// Your data here.
	requestTab map[int64]requestEntry
	chanMap    map[int]chan Config
	configs    []Config // indexed by config num
}
type requestEntry struct {
	RpcID  int
	Config Config
}

type Op struct {
	// Your data here.
	Optype    string //  Join / Leave / Move / Query
	Servers   map[int][]string
	GID       int
	GIDs      []int
	Shard     int
	ConfigNum int
	ClientID  int64
	RpcID     int
}

func (sc *ShardCtrler) Join(args *JoinArgs, reply *JoinReply) {
	// Your code here.
	// 检查leader
	if _, isLeader := sc.rf.GetState(); !isLeader {
		reply.WrongLeader = true
		return
	}

	// 重复检测
	if ok, _ := sc.checkRpc(args.ClientId, args.RpcId); ok {
		DPrintf("重复RPC: %v", args.RpcId)
		reply.WrongLeader = false
		return
	}

	// start 日志添加
	timer := time.NewTimer(JoinOutTime)

	option := Op{
		Optype:   JoinType,
		Servers:  args.Servers,
		ClientID: args.ClientId,
		RpcID:    args.RpcId,
	}

	index, _, isLeader := sc.rf.Start(option)
	if !isLeader {
		reply.WrongLeader = true
		return
	}

	// 等待apply后回复

	ch := sc.getOrMakeCh(index)
	DPrintf("////{S%v}等待一致性通过或超时,Optype: Join,{C%v},RpcID: %v,servers[%v]", sc.me, args.ClientId, args.RpcId, args.Servers)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时Join, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = true

	case <-ch:
		DPrintf("{S%v} 成功Join, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = false
	}

	timer.Stop()
	sc.deleteCh(index)

}

func (sc *ShardCtrler) Leave(args *LeaveArgs, reply *LeaveReply) {
	// Your code here.
	if _, isLeader := sc.rf.GetState(); !isLeader {
		reply.WrongLeader = true
		return
	}

	// 重复检测
	if ok, _ := sc.checkRpc(args.ClientId, args.RpcId); ok {
		DPrintf("重复RPC: %v", args.RpcId)
		reply.WrongLeader = false
		return
	}

	// start 日志添加
	timer := time.NewTimer(LeaveOutTime)

	option := Op{
		Optype:   LeaveType,
		GIDs:     args.GIDs,
		ClientID: args.ClientId,
		RpcID:    args.RpcId,
	}

	index, _, isLeader := sc.rf.Start(option)
	if !isLeader {
		reply.WrongLeader = true
		return
	}

	// 等待apply后回复

	ch := sc.getOrMakeCh(index)
	DPrintf("////{S%v}等待一致性通过或超时,Optype: Leave,{C%v},RpcID: %v,leaveGIDs[%v]", sc.me, args.ClientId, args.RpcId, args.GIDs)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时Leave, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = true

	case <-ch:
		DPrintf("{S%v} 成功Leave, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = false
	}

	timer.Stop()
	sc.deleteCh(index)

}

func (sc *ShardCtrler) Move(args *MoveArgs, reply *MoveReply) {
	// Your code here.
	if _, isLeader := sc.rf.GetState(); !isLeader {
		reply.WrongLeader = true
		return
	}

	// 重复检测
	if ok, _ := sc.checkRpc(args.ClientId, args.RpcId); ok {
		DPrintf("重复RPC: %v", args.RpcId)
		reply.WrongLeader = false
		return
	}

	// start 日志添加
	timer := time.NewTimer(MoveOutTime)

	option := Op{
		Optype:   MoveType,
		Shard:    args.Shard,
		GID:      args.GID,
		ClientID: args.ClientId,
		RpcID:    args.RpcId,
	}

	index, _, isLeader := sc.rf.Start(option)
	if !isLeader {
		reply.WrongLeader = true
		return
	}

	// 等待apply后回复

	ch := sc.getOrMakeCh(index)
	DPrintf("////{S%v}等待一致性通过或超时,Optype: Move,{C%v},RpcID: %v", sc.me, args.ClientId, args.RpcId)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时Move, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = true

	case <-ch:
		DPrintf("{S%v} 成功Move, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = false
	}

	timer.Stop()
	sc.deleteCh(index)

}

func (sc *ShardCtrler) Query(args *QueryArgs, reply *QueryReply) {
	// Your code here.
	if _, isLeader := sc.rf.GetState(); !isLeader {
		reply.WrongLeader = true
		return
	}

	// 重复检测
	if ok, _ := sc.checkRpc(args.ClientId, args.RpcId); ok {
		DPrintf("重复RPC: %v", args.RpcId)
		reply.WrongLeader = false
		return
	}

	// start 日志添加
	timer := time.NewTimer(QueryOutTime)

	option := Op{
		Optype:    QueryType,
		ConfigNum: args.Num,
		ClientID:  args.ClientId,
		RpcID:     args.RpcId,
	}

	index, _, isLeader := sc.rf.Start(option)
	if !isLeader {
		reply.WrongLeader = true
		return
	}

	// 等待apply后回复

	ch := sc.getOrMakeCh(index)
	DPrintf("////{S%v}等待一致性通过或超时,Optype: Query,{C%v},RpcID: %v", sc.me, args.ClientId, args.RpcId)

	select {
	case <-timer.C:
		DPrintf("{S%v} 超时Query, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = true

	case config_query := <-ch:
		DPrintf("{S%v} 成功Query, {C%v} rpcId: %v", sc.me, args.ClientId, args.RpcId)
		reply.WrongLeader = false
		reply.Config = config_query
	}

	timer.Stop()
	sc.deleteCh(index)

}

// the tester calls Kill() when a ShardCtrler instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
func (sc *ShardCtrler) Kill() {
	sc.rf.Kill()
	// Your code here, if desired.
}

// needed by shardkv tester
func (sc *ShardCtrler) Raft() *raft.Raft {
	return sc.rf
}

// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant shardctrler service.
// me is the index of the current server in servers[].
func StartServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister) *ShardCtrler {
	sc := new(ShardCtrler)
	sc.me = me

	sc.configs = make([]Config, 1)
	sc.configs[0].Groups = map[int][]string{}
	sc.configs[0].Num = 0
	// sc.configs[0].Shards = make([]int, NShards)
	//sc.configNum = 0

	labgob.Register(Op{})
	sc.applyCh = make(chan raft.ApplyMsg, 1)
	sc.rf = raft.Make(servers, me, persister, sc.applyCh)

	// Your code here.
	//sc.severNum = len(servers)
	sc.requestTab = make(map[int64]requestEntry)
	sc.chanMap = make(map[int]chan Config)

	go sc.receiver()

	return sc
}

func (sc *ShardCtrler) reBalance(gids []int) [NShards]int {
	DPrintf("{sc%v} reBalance,gids[%v]", sc.me, gids)
	serversNum := len(gids)
	if serversNum == 0 {
		return [NShards]int{}
	}

	average := NShards / serversNum
	left := NShards % serversNum // 余数，最多left个sever能取到average+1
	newShards := sc.configs[len(sc.configs)-1].Shards

	// 初始状态，没有server
	if len(sc.configs) == 1 {

		// 均匀分配至average
		for i := 0; i < serversNum; i++ {
			for j := 0; j < average; j++ {
				newShards[j*serversNum+i] = gids[i]
			}
		}
		// 对余数进行分配
		for i := 0; i < left; i++ {
			newShards[average*serversNum+i] = gids[i]
		}
		DPrintf("{sc%v} reBalance初始化,newShards[%v]", sc.me, newShards)
		return newShards

	} else {
		// 修改配置后的均衡
		var temp_shards []int           // 未分配的shards集合
		num_pergid := make(map[int]int) // 当前每个gid对应的切片数量
		oldShards := sc.configs[len(sc.configs)-1].Shards

		// 先获取之前的gid的切片数量
		for i := 0; i < NShards; i++ {
			is_add := false
			for _, newgid := range gids {
				if oldShards[i] == newgid {
					num_pergid[newgid]++
					is_add = true
				}
			}
			// 未分配的shards
			if !is_add {
				temp_shards = append(temp_shards, i)
			}
		}

		// 删减 如果有余数，容许left个sever达到averge+1，否则容许达到average
		for _, gid := range gids {
			if left > 0 {
				if num_pergid[gid] > average+1 {
					left--
					for j, u := 0, 0; j < NShards; j++ {
						// 生成新的shards数组，某个gid组多只能包含average+1个,多出来的放到未分配数组中
						if oldShards[j] == gid && u < average+1 {
							newShards[j] = gid
							u++
						} else if oldShards[j] == gid && u == average+1 {
							temp_shards = append(temp_shards, j)
						}
					}
					num_pergid[gid] = average + 1
				}
			} else {
				if num_pergid[gid] > average {
					for j, u := 0, 0; j < NShards; j++ {
						// 生成新的shards数组，某个gid组多只能包含average+1个,多出来的放到未分配数组中
						if oldShards[j] == gid && u < average {
							newShards[j] = gid
							u++
						} else if oldShards[j] == gid && u == average {
							temp_shards = append(temp_shards, j)
						}
					}
					num_pergid[gid] = average
				}
			}
		}
		DPrintf("{sc%v} reBalance删减完成,newShards[%v],temp_shards[%v]", sc.me, newShards, temp_shards)
		// 增添到averge 刚好到temp_shards用完
		for u, j := 0, 0; u < serversNum && j < len(temp_shards); u++ {
			for num_pergid[gids[u]] < average && j < len(temp_shards) {
				newShards[temp_shards[j]] = gids[u]
				j++
				num_pergid[gids[u]]++
			}
		}
		DPrintf("{sc%v} reBalance增添完成,newShards[%v],temp_shards[%v]", sc.me, newShards, temp_shards)
	}

	return newShards
}

func (sc *ShardCtrler) apply(recvOp Op) Config {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	commentRet := Config{}
	// 检查是否重复执行
	if request_entry, ok := sc.requestTab[recvOp.ClientID]; ok && recvOp.RpcID == request_entry.RpcID {
		// 已有的条目 直接返回结果
		DPrintf("!!!! {S%v} cmd already apply Optype: %v,{C%v},RpcID: %v", sc.me, recvOp.Optype, recvOp.ClientID, recvOp.RpcID)
		if recvOp.Optype == QueryType {
			commentRet = request_entry.Config
		}
		return commentRet
	}

	// 执行新的条目
	switch recvOp.Optype {
	case JoinType:

		// 当前config
		lastConfig := sc.configs[len(sc.configs)-1]

		// 新的config生成
		newGroup := make(map[int][]string)
		var newGids []int

		for gid, servers := range lastConfig.Groups {
			newGroup[gid] = servers
			newGids = append(newGids, gid)
		}

		for gid, servers := range recvOp.Servers {
			newGroup[gid] = servers
			newGids = append(newGids, gid)
		}
		// 保证配置的确定性
		sort.Ints(newGids)
		newShards := sc.reBalance(newGids)

		// 修改状态机为新的config
		newConfig := Config{
			Num:    lastConfig.Num + 1,
			Shards: newShards,
			Groups: newGroup,
		}
		sc.configs = append(sc.configs, newConfig)
		DPrintf("{sc%v} Join apply,shards[%v]", sc.me, newShards)
	case LeaveType:
		// 当前config
		lastConfig := sc.configs[len(sc.configs)-1]

		// 新的config生成
		newGroup := make(map[int][]string)
		var newGids []int

		for gid, servers := range lastConfig.Groups {
			// 不在leave的GIDs中
			is_leave := false
			for _, leavegid := range recvOp.GIDs {
				if gid == leavegid {
					is_leave = true
				}
			}
			if !is_leave {
				newGroup[gid] = servers
				newGids = append(newGids, gid)
			}
		}
		// 保证配置的确定性
		sort.Ints(newGids)
		newShards := sc.reBalance(newGids)

		// 修改状态机为新的config
		newConfig := Config{
			Num:    lastConfig.Num + 1,
			Shards: newShards,
			Groups: newGroup,
		}
		sc.configs = append(sc.configs, newConfig)
		DPrintf("{sc%v} Leave apply,shards[%v]", sc.me, newShards)
	case MoveType:
		// 当前config
		lastConfig := sc.configs[len(sc.configs)-1]

		// 新的config生成
		newGroup := make(map[int][]string)
		//var newGids []int

		for gid, servers := range lastConfig.Groups {
			newGroup[gid] = servers
			//newGids = append(newGids, gid)
		}

		newShards := lastConfig.Shards

		newShards[recvOp.Shard] = recvOp.GID

		// 修改状态机为新的config
		newConfig := Config{
			Num:    lastConfig.Num + 1,
			Shards: newShards,
			Groups: newGroup,
		}
		sc.configs = append(sc.configs, newConfig)
		DPrintf("{sc%v} Move apply,shards[%v]", sc.me, newShards)
	case QueryType:
		// do nothing
	default:
		DPrintf("{sc} undeclared RPC type")
	}

	// QueryType
	newEntry := requestEntry{
		RpcID: recvOp.RpcID,
	}
	if recvOp.Optype == QueryType {
		if recvOp.ConfigNum != -1 && recvOp.ConfigNum < len(sc.configs) {
			newEntry.Config = sc.configs[recvOp.ConfigNum]
		} else {
			newEntry.Config = sc.configs[len(sc.configs)-1]
		}
		commentRet = newEntry.Config
		DPrintf("{sc%v} Query apply,config[%v]", sc.me, newEntry.Config)
	}
	// 更新表
	sc.requestTab[recvOp.ClientID] = newEntry
	DPrintf("{S%v}更新表kv.requestTab[%v] latest RPCid: %v ", sc.me, recvOp.ClientID, newEntry.RpcID)
	return commentRet
}

func (sc *ShardCtrler) receiver() {
	for {
		select {
		case recv_msg := <-sc.applyCh:
			if recv_msg.CommandValid {
				DPrintf("{sc%v} receive Op,index[%v]", sc.me, recv_msg.CommandIndex)

				recvOp := recv_msg.Command.(Op)
				DPrintf("{sc%v} recvOp: [%v]", sc.me, recvOp.Optype)

				// 应用
				commentRet := sc.apply(recvOp)

				// 唤醒 返回结果 只有leader在处理RPC,只有通道还没被删除的时候需要返回
				if _, isleader := sc.rf.GetState(); isleader {
					if ch, ok := sc.getChIfHas(recv_msg.CommandIndex); ok {
						ch <- commentRet
					}
				}

			}
		}

	}
}

// 重复检验，查找client对应的rpc是否已经执行
func (sc *ShardCtrler) checkRpc(clientId int64, rpcId int) (bool, Config) {
	// 检查表
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// 已经执行RPC请求
	equset_entry, ok := sc.requestTab[clientId]
	if ok && equset_entry.RpcID == rpcId {
		return true, equset_entry.Config
	}
	return false, Config{}
}

func (sc *ShardCtrler) getOrMakeCh(index int) chan Config {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, ok := sc.chanMap[index]; !ok {
		sc.chanMap[index] = make(chan Config, 1)
		DPrintf("{S%v}创建chan[index: %v]", sc.me, index)
	}

	return sc.chanMap[index]
}

func (sc *ShardCtrler) getChIfHas(index int) (chan Config, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	ch, ok := sc.chanMap[index]
	if !ok {
		DPrintf("{S%v}不存在chan[index: %v]", sc.me, index)
	}
	return ch, ok
}

func (sc *ShardCtrler) deleteCh(index int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.chanMap, index)
}
