package shardctrler

import "time"

//
// Shard controler: assigns shards to replication groups.
//
// RPC interface:
// Join(servers) -- add a set of groups (gid -> server-list mapping).
// Leave(gids) -- delete a set of groups.
// Move(shard, gid) -- hand off one shard from current owner to gid.
// Query(num) -> fetch Config # num, or latest config if num==-1.
//
// A Config (configuration) describes a set of replica groups, and the
// replica group responsible for each shard. Configs are numbered. Config
// #0 is the initial configuration, with no groups and all shards
// assigned to group 0 (the invalid group).
//
// You will need to add fields to the RPC argument structs.
//

// The number of shards.
const NShards = 10

// A configuration -- an assignment of shards to groups.
// Please don't change this.
type Config struct {
	Num    int              // config number 配置id
	Shards [NShards]int     // shard -> gid 表示每个切片素在的group
	Groups map[int][]string // gid -> servers[] 表示每个group包含的servers
}

const (
	OK = "OK"

	JoinType  = "Join"
	LeaveType = "Leave"
	MoveType  = "Move"
	QueryType = "Query"

	JoinOutTime  = 500 * time.Millisecond
	LeaveOutTime = 500 * time.Millisecond
	MoveOutTime  = 500 * time.Millisecond
	QueryOutTime = 500 * time.Millisecond
)

type Err string

type JoinArgs struct {
	Servers  map[int][]string // new GID -> servers mappings
	ClientId int64
	RpcId    int
}

type JoinReply struct {
	WrongLeader bool
	Err         Err
}

type LeaveArgs struct {
	GIDs     []int
	ClientId int64
	RpcId    int
}

type LeaveReply struct {
	WrongLeader bool
	Err         Err
}

type MoveArgs struct {
	Shard    int
	GID      int
	ClientId int64
	RpcId    int
}

type MoveReply struct {
	WrongLeader bool
	Err         Err
}

type QueryArgs struct {
	Num      int // desired config number
	ClientId int64
	RpcId    int
}

type QueryReply struct {
	WrongLeader bool
	Err         Err
	Config      Config
}
