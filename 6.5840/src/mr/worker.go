package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	// 生成哈希值
	h := fnv.New32a()
	// 写入字节数
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	//CallExample()

	for {
		args := TaskRequest{}
		reply := TaskResponse{}

		CallGetTask(&args, &reply)
		state := reply.State

		if state == 0 && reply.Xtask.MapId != -1 {
			//fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>> map task\n")
			filename := reply.Xtask.FileName
			id := strconv.Itoa(reply.Xtask.MapId)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			content, _ := ioutil.ReadAll(file)
			file.Close()
			// 调用map函数
			kva := mapf(filename, string(content))

			// 按照reduce数量分桶
			num_reduce := reply.NumReduceTask
			bucket := make([][]KeyValue, num_reduce)
			//生成hash值 填桶
			for _, kv := range kva {
				num := ihash(kv.Key) % num_reduce
				bucket[num] = append(bucket[num], kv)
			}
			// 写中间文件
			for i := 0; i < num_reduce; i++ {
				tmp_file, err := ioutil.TempFile("", "mr-map-*") // 创建临时文件 第一个参数是目录，第二个是文件名
				if err != nil {
					log.Fatalf("cannot create tmp_file")
				}
				// 存储kv在中间文件 JSON编码
				enc := json.NewEncoder(tmp_file) // 创建JSON编辑器，用于写入tmp_file

				err = enc.Encode(bucket[i]) // 将buckt[i]的值编码为JSON格式，写入文件
				if err != nil {
					log.Fatalf("encode bucket error")
				}

				tmp_file.Close()
				out_file_name := "mr-" + id + "-" + strconv.Itoa(i) // mr-X-Y
				os.Rename(tmp_file.Name(), out_file_name)
			}

			CallTaskFin(&reply.Xtask)
		} else if state == 1 && reply.Xtask.ReduceId != -1 {
			// reduce

			//fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>> reduce task\n")
			id := strconv.Itoa(reply.Xtask.ReduceId)
			num_map := reply.NumMapTask
			intermediate := []KeyValue{}
			for i := 0; i < num_map; i++ {
				map_filename := "mr-" + strconv.Itoa(i) + "-" + id
				inputfiles, err := os.OpenFile(map_filename, os.O_RDONLY, 0777)
				if err != nil {
					log.Fatalf("cannot open redceTask %v", map_filename)
				}
				dec := json.NewDecoder(inputfiles)
				for {
					var kv []KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					intermediate = append(intermediate, kv...) // ...表示把切片打散成元素后加入
				}
			}

			sort.Sort(ByKey(intermediate))
			// reduce 输出文件
			tmp_file, err := ioutil.TempFile("", "mr-reduce-*")
			if err != nil {
				log.Fatalf("cannot create tmp_file")
			}
			// 去重 并运行reduce
			i := 0
			for i < len(intermediate) {
				j := i + 1
				//相同key一起处理
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(tmp_file, "%v %v\n", intermediate[i].Key, output)

				i = j
			}
			tmp_file.Close()
			out_file_name := "mr-out-" + id
			os.Rename(tmp_file.Name(), out_file_name)

			CallTaskFin(&reply.Xtask)

		} else if state == 2 {
			fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>> Fin\n")
			break
		} else {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(time.Second)
	}
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
// func CallExample() {

// 	// declare an argument structure.
// 	args := ExampleArgs{}

// 	// fill in the argument(s).
// 	args.X = 99

// 	// declare a reply structure.
// 	reply := ExampleReply{}

// 	// send the RPC request, wait for the reply.
// 	// the "Coordinator.Example" tells the
// 	// receiving server that we'd like to call
// 	// the Example() method of struct Coordinator.
// 	ok := call("Coordinator.Example", &args, &reply)
// 	if ok {
// 		// reply.Y should be 100.
// 		fmt.Printf("reply.Y %v\n", reply.Y)
// 	} else {
// 		fmt.Printf("call failed!\n")
// 	}
// }

func CallGetTask(args *TaskRequest, reply *TaskResponse) {

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetTask", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("call get task ok!\n")
	} else {
		fmt.Printf("gettask call failed!\n")
	}
}

func CallTaskFin(args *Task) {

	// ok := call("Coordinator.TaskFin", nil, nil)
	reply := ExampleReply{}

	ok := call("Coordinator.TaskFin", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("call task fin ok!\n")
	} else {
		fmt.Printf("taskfin call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// 连接到指定网络地址的HTTP RPC服务器，监听
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	// 参数是指定协议额和服务器地址的字符串，返回客户端对象
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()
	// 在单独的goroutine中发送RPC请求，异步调用
	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
