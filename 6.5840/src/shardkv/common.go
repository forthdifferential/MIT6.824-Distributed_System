package shardkv

import "time"

//
// Sharded key/value server.
// Lots of replica groups, each running Raft.
// Shardctrler decides which group serves each shard.
// Shardctrler may change shard assignment from time to time.
//
// You will have to modify these definitions.
//

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongGroup  = "ErrWrongGroup"
	ErrWrongLeader = "ErrWrongLeader"
	ErrTimeOut     = "ErrTimeOut"

	GetOutTime       = 500 * time.Millisecond
	PutAppendOutTime = 500 * time.Millisecond

	// shard状态
	StateOk   = "OK"
	StateOut  = "Out"  // 等待移走
	StateIn   = "In"   // 等待移入
	StateWait = "Wait" // 移入完成，等待告知ex-ower删除

	ErrConfigNumNotMatch = "ConfigNumNotMatch"
	ErrConfigNumNotReady = "ConfigNumNotReady"
	ErrConfigNumOut      = "ErrConfigNumOut"
)

type Err string

// Put or Append
type PutAppendArgs struct {
	// You'll have to add definitions here.
	Key    string
	Value  string
	Optype string // "Put" or "Append"

	ClientId int64
	RpcId    int
}

type PutAppendReply struct {
	Err Err
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	ClientId int64
	RpcId    int
}

type GetReply struct {
	Err   Err
	Value string
}
