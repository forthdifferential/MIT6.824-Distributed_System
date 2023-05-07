package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"sync"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type TaskRequest struct {
	X int
}
type TaskResponse struct {
	Xtask         Task
	NumMapTask    int
	NumReduceTask int
	State         int32
}

func lenTaskFin(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		if v.(TimeStamp).Fin {
			i++
		}
		return true
	})
	// for  _, j range m {
	// 	if m.(TimeStamp).Fin {
	// 		i++
	// 	}
	// }
	return i
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	// 用户ID 转化为字符串表示 加到s后面
	s += strconv.Itoa(os.Getuid())
	return s
}
