package kvraft

import (
	"crypto/rand"
	"math/big"

	"6.5840/labrpc"
)

type Clerk struct {
	servers   []*labrpc.ClientEnd
	serverNum int
	// You will have to modify this struct.
	requestNum int
	clientId   int64
	leaderId   int
}

// 生成真随机数上限是2的62次
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.serverNum = len(servers)
	// You'll have to add code here.
	ck.requestNum = 0
	ck.leaderId = 0
	ck.clientId = nrand()
	return ck
}

// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
// 获取k对应的v
func (ck *Clerk) Get(key string) string {

	// You will have to modify this function.

	ck.requestNum++
	serverId := ck.leaderId

	args := GetArgs{
		Key:      key,
		ClientID: ck.clientId,
		RpcID:    ck.requestNum,
	}
	reply := GetReply{}
	DPrintf("++++{C%v} Get,RpcID: %v,try,key: %v", args.ClientID, args.RpcID, args.Key)
	ok := ck.servers[serverId].Call("KVServer.Get", &args, &reply)

	// keeps trying forever in the face of all other errors.
	for !(ok && reply.Success) {
		// 如果RPC无法到达server，发送其他server来重试

		//DPrintf("CLERK Get {client %v} fail, key: %v", serverId, args.Key)
		//oldserver := serverId
		serverId = (serverId + 1) % ck.serverNum

		//DPrintf("{C%v} Get{oldS%v} error:%v,retry{server %v},rpcId: %v,key: %v", args.ClientID, oldserver, reply.Err, serverId, args.RpcID, args.Key)
		DPrintf("{C%v} Get error:%v,retry,rpcId: %v,key: %v", args.ClientID, reply.Err, args.RpcID, args.Key)

		reply = GetReply{}
		ok = ck.servers[serverId].Call("KVServer.Get", &args, &reply)

	}

	DPrintf("----{C%v} Get {S%v} rpcId: %v,finish", args.ClientID, serverId, args.RpcID)
	ck.leaderId = serverId // 成功发送的一定是leader所在server

	return reply.Value
}

// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) PutAppend(key string, value string, op string) {
	// You will have to modify this function.
	ck.requestNum++
	serverId := ck.leaderId

	args := PutAppendArgs{
		Key:      key,
		Value:    value,
		Optype:   op,
		ClientID: ck.clientId,
		RpcID:    ck.requestNum,
	}
	reply := PutAppendReply{}
	// DPrintf("++++{C%v} PutAppend {S%v},RpcID: %v,try type: %v,key: %v", args.ClientID, serverId, args.RpcID, args.Optype, args.Key)
	DPrintf("++++{C%v} PutAppend ,RpcID: %v,try type: %v,key: %v", args.ClientID, args.RpcID, args.Optype, args.Key)
	ok := ck.servers[serverId].Call("KVServer.PutAppend", &args, &reply)

	// keeps trying forever in the face of all other errors.
	for !(ok && reply.Success) {
		// !ok表示RPC超时，发送其他server来重试
		// DPrintf("CLERK PutAppend {client %v} fail: %v,key: %v,value: %v", serverId, args.Optype, args.Key, args.Value)
		// oldserver := serverId
		serverId = (serverId + 1) % ck.serverNum

		// DPrintf("{C%v} PutAppend{oldS%v} error: %v,retry {S%v} type: %v,key: %v", args.ClientID, oldserver, reply.Err, serverId, args.Optype, args.Key)
		DPrintf("{C%v} PutAppend error: %v,retry type: %v,key: %v", args.ClientID, reply.Err, args.Optype, args.Key)

		reply = PutAppendReply{}
		ok = ck.servers[serverId].Call("KVServer.PutAppend", &args, &reply)
	}

	// 如果RPC无法到达server，发送其他server来重试
	DPrintf("----{C%v} PutAppend {S%v} rpcId: %v,finish", args.ClientID, serverId, args.RpcID)
	ck.leaderId = serverId

}

// 替换k对应的v
func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}

// 追加k对应的v
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
