package kvraft

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
)

type Err string

// Put or Append
type PutAppendArgs struct {
	Key    string
	Value  string
	Optype string // "Put" or "Append"
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	ClientID int64
	RpcID    int
}

type PutAppendReply struct {
	Err     Err
	Success bool
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	ClientID int64
	RpcID    int
}

type GetReply struct {
	Err     Err
	Value   string
	Success bool
}
