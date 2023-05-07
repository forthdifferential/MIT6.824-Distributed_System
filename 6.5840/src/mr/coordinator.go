package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Coordinator struct {
	// Your definitions here.
	State          int32 // 0map 1reduce 2fin
	NumMapTask     int
	NumReduceTask  int
	MapTask        chan Task
	MapTaskTime    sync.Map // 并发安全的map
	ReduceTask     chan Task
	ReduceTaskTime sync.Map
	files          []string
}
type TimeStamp struct {
	Time int64
	Fin  bool
}

type Task struct {
	FileName string
	MapId    int
	ReduceId int
	//State    int // 0start 1running 2finish
}

var mu sync.Mutex

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
// func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
// 	reply.Y = args.X + 1
// 	return nil
// }

func (c *Coordinator) GetTask(args *TaskRequest, reply *TaskResponse) error {
	mu.Lock()
	defer mu.Unlock()
	state := atomic.LoadInt32(&c.State)
	// time_now := time.Now().Unix()
	switch state {
	case 0:
		{
			if len(c.MapTask) != 0 {
				maptask, ok := <-c.MapTask
				if ok {
					reply.Xtask = maptask
					// 1. 能不能改chan里面的元素的time 2.或者在get的时候才设置时间 不太对 需要提前存储
					// timIF, _ = c.MapTaskTime.Load(maptask.MapId)
					// maptask
					//
					fmt.Printf("Got 1 map task\n")
				}
			} else {
				reply.Xtask.MapId = -1
			}
		}
	case 1:
		{
			if len(c.ReduceTask) != 0 {
				reducetask, ok := <-c.ReduceTask
				if ok {
					reply.Xtask = reducetask
					fmt.Printf("Got 1 reduce task\n")
				}
			} else {
				reply.Xtask.ReduceId = -1
			}
		}
	}
	reply.NumMapTask = c.NumMapTask
	reply.NumReduceTask = c.NumReduceTask
	reply.State = c.State

	return nil
}

func (c *Coordinator) TaskFin(args *Task, reply *ExampleReply) error {
	mu.Lock()
	defer mu.Unlock()
	time_now := time.Now().Unix()
	if lenTaskFin(&c.MapTaskTime) != c.NumMapTask {

		start_timeIF, _ := c.MapTaskTime.Load(args.MapId)
		if time_now-start_timeIF.(TimeStamp).Time > 10 {
			return nil
		}

		c.MapTaskTime.Store(args.MapId, TimeStamp{time_now, true})
		if lenTaskFin(&c.MapTaskTime) == c.NumMapTask {
			atomic.StoreInt32(&c.State, 1) // 原子化修改，因为有race
			// map任务结束之后 填充reduce任务
			for i := 0; i < c.NumReduceTask; i++ {
				c.ReduceTask <- Task{ReduceId: i}
				c.ReduceTaskTime.Store(i, TimeStamp{time_now, false})
			}
		}
	} else if lenTaskFin(&c.ReduceTaskTime) != c.NumReduceTask {
		start_timeIF, _ := c.ReduceTaskTime.Load(args.ReduceId)
		if time_now-start_timeIF.(TimeStamp).Time > 10 {
			return nil
		}
		c.ReduceTaskTime.Store(args.ReduceId, TimeStamp{time_now, true})
		if lenTaskFin(&c.ReduceTaskTime) == c.NumReduceTask {
			atomic.StoreInt32(&c.State, 2)
		}
	}

	return nil
}

func (c *Coordinator) TimeTick() {
	state := atomic.LoadInt32(&c.State)
	time_now := time.Now().Unix()

	if state == 0 {
		for i := 0; i < c.NumMapTask; i++ {
			tmp, _ := c.MapTaskTime.Load(i)
			if !tmp.(TimeStamp).Fin && time_now-tmp.(TimeStamp).Time > 10 {
				fmt.Println("map time out!")
				c.MapTask <- Task{FileName: c.files[i], MapId: i} //TODO 这里硬加Task没有去掉原来的Task 可能会导致跑很多次
				c.MapTaskTime.Store(i, TimeStamp{time_now, false})
			}
		}
	} else if state == 1 {
		for i := 0; i < c.NumReduceTask; i++ {
			tmp, _ := c.ReduceTaskTime.Load(i)
			if !tmp.(TimeStamp).Fin && time_now-tmp.(TimeStamp).Time > 10 {
				fmt.Println("reduce time out!")
				c.ReduceTask <- Task{ReduceId: i}
				c.ReduceTaskTime.Store(i, TimeStamp{time_now, false})
			}
		}
	}
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {

	c.TimeTick()

	ret := false
	// Your code here.
	state := atomic.LoadInt32(&c.State)
	if state == 2 {
		ret = true
	}
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		State:         0,
		NumMapTask:    len(files),
		NumReduceTask: nReduce,
		MapTask:       make(chan Task, len(files)),
		ReduceTask:    make(chan Task, nReduce),
		files:         files,
	}

	// Your code here.
	time_now := time.Now().Unix()
	for i, file := range files {
		c.MapTask <- Task{FileName: file, MapId: i}
		c.MapTaskTime.Store(i, TimeStamp{time_now, false})
	}

	c.server()
	return &c
}
