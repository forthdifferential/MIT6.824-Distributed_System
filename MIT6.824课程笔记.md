# MIT6.824课程笔记

## introduction

分布式：通过网络来协调，协同完成一致任务的一些计算机。

分布式系统
一组不共享内存和时钟处理器的集合。每个处理器都有自己的内存，之间通过各种通信网络加以实现。处理器大小和和功能不同，其中，某个站点表示特定的主机，服务器，拥有客户机所需的资源。

### 建立分布式的原因

1. 加快计算速度：将特定的计算分化成可以并发运行的子运算，分布式系统允许将这些子运算分不到不同的站点，就可以并发的进行。此外啊，还可以进行负载分配。（大量的计算机意味着）大量的并行运算，大量CPU、大量内存、以及大量磁盘在并行的运行
2. 可靠性（提供容错）：如果系统具有足够的冗余（包括硬件和数据），即使某些站点出现故障，系统也会继续运行下去。1. 分布式系统必须保证能检测到站点故障，不再使用该站点的服务，并采取适当措施修复；2. 如果有其他站点可以接管，系统必须保证操作转移的正确；3. 故障的站点恢复后，系统保证有一个顺利结合到系统的机制；
3. 一些问题天然在空间上是分布的。例如银行转账，我们假设银行A在纽约有一台服务器，银行B在伦敦有一台服务器，这就需要一种两者之间协调的方法。所以，有一些天然的原因导致系统是物理分布的。
4. 安全，通过隔离提高安全性。比如一些代码不被信任但需要交互，可以将代码分散多出运行，系统分成多个计算机，这样可以限制出错域。

### 分布式难点：

1. 并发编程
2. 复杂交互，时间依赖（同步、异步）
3. 局部错误很难调试
4. 很难获得高性能

### 基础架构： 存储 通信 计算

存储是最重要，构建一种多副本、容错的、高性能分布式存储实现；

存储和计算，目的是实现简单剪口，像非分布式存储和计算系统一样被第三方应用，但是实际上又是一个有极高的性能和容错的分布式系统。

**构建用到的工具：**

RPC（Remote Procedure Call）。RPC的目标就是掩盖我们正在不可靠网络上通信的事实。一种分布式系统中的通信方式，可以使得不同的计算机程序之间可以像本地程序一样进行相互调用，从而方便构建分布式系统。通过RPC技术，可以帮助实现以下功能 ：

1. 分布式服务调用：各个计算机之间可以调用对方提供的服务，从而实现分布式系统中的协同工作。
2. 负载均衡：可以将请求分配到多个服务器上，从而实现负载均衡，提高系统的可靠性和可用性。
3. 高并发处理：，可以实现多个请求的并发处理，提高系统的响应速度。
4. 数据共享：可以将数据共享到多个计算机之间，从而实现数据共享和协同处理。

多线程编程技术

考虑控制并发、锁

### 拓展性Sacalability

如果我用一台计算机解决了一些问题，当我买了第二台计算机，我只需要一半的时间就可以解决这些问题，或者说每分钟可以解决两倍数量的问题。两台计算机构成的系统如果有两倍性能或者吞吐，就是我说的可扩展性。

通过增加机器的方式来实现扩展，但是现实中这很难实现，需要一些架构设计来将这个可扩展性无限推进下去。

比如说并发量上去了，那需要增加服务器数量，这导致数据库成为瓶颈，又需要搭建分布式存储：
		也就是web服务器和数据库的成为瓶颈时，都能突破，拓展下去。

### 可用 | 容错性Avaliability

可用性（Availability）: 这样在特定的错误类型下，系统仍然能够正常运行，仍然可以像没有出现错误一样，为你提供完整的服务。

大型分布式系统中有一个大问题，那就是一些很罕见的问题会被放大,各个地方总是有一些小问题出现。所以大规模系统会将一些几乎不可能并且你不需要考虑的问题，变成一个持续不断的问题。

自我可恢复性（recoverability）: 如果出现了问题，服务会停止工作(通常需要做一些操作，例如将最新的数据存放在磁盘中),不再响应请求，之后有人来修复，并且在修复之后系统仍然可以正常运行，就像没有出现过问题一样

**可用和自我恢复用到的工具**

非易失存储：存放一些checkpoint或者系统状态的log在这些存储中，这样当备用电源恢复或者某人修好了电力供给，我们还是可以从硬盘中读出系统最新的状态，并从那个状态继续运行。但是非易失存储的读写的代价很高，避免频繁读写

复制：管理多副本实现容错系统

### 一致性Consistency

定义操作行为的概念。因为分布式系统考虑性能和容错，通常会使用多个副本有不同版本的 key-value 。比如在put操作中，有两个服务器，需要发给两个副本以达到同步，但是发送过程中可能客户端故障了，只有一个收到了，数据不一致了。

弱一致是指，不保证get请求可以得到最近一次完成的put请求写入的值。尽管有很多细节的工作要处理，强一致可以保证get得到的是put写入的最新的数据。所以在一个弱一致系统中，某人通过put请求写入了一个数据，但是你通过get看到的可能仍然是一个旧数据，而这个旧数据可能是很久之前写入的。  但是强一致的开销可能非常大，常常构建弱一致系统，只从最近的数据副本更新和获取数据。

### 提高性能 performance

- throughput 吞吐量：目标是吞吐量与部署的机器数成正比
- latency 低延迟：其中一台机器执行慢就会导致整个请求响应慢，这称为**尾部延迟(tail latency)**

### MapReduce基本工作方式

MapReduce的思想是，应用程序设计人员和分布式运算的使用者，只需要写简单的Map函数和Reduce函数，而不需要知道任何有关分布式的事情，MapReduce框架会处理剩下的事情。

 map-reduce经典举例即统计字母出现的次数，多个进程各自通过map函数统计获取到的数据片段的字母的出现次数；后续再通过reduce函数，汇总聚合map阶段下每个进程对各自负责的数据片段统计的字母出现次数。一旦执行了shuffle，多个reduce函数可以各自只聚合一种字母的出现总次数，彼此之间不干扰。

开销昂贵的部分即shuffle，map的结果经过shuffle按照一定的顺序整理/排序，然后才分发给不同的reduce处理。这里shuffle的操作理论比map、reduce昂贵。

## MapReduce

输入一组kv，生成一组kv。

MAP函数获取一个输入，并且生成一组中间的KV，

MapReduce库把与同一中间键i相关联的所有中间值组合在一起，通过迭代器传递给Reduce函数。

Reduce函数接受中间键i和键对应的一组值，把值合在一起，形成一个看更小的值集。一般每次只生成0或1个输出值。

<img src="D:\MyTxt\typoraPhoto\image-20230329113648794.png" alt="image-20230329113648794" style="zoom:80%;" />

### 步骤：

1. 输入文件，MapReduce库分割成M个片段，每个片段16-64MB，然后再一群机器上启动程序的副本
2. 其中一个副本是Master，其余是workers，总共有M个map任务和R个reduce任务需要分配。由主进程挑选空闲的workers分配一项任务。
3. 分配到map任务的worker读取内容（被拆分好的一部分），从输出数据中解析出KV，传递给Map函数，然后把生成的中间KV存在内存
4. 定期写入本地磁盘，并分区R个区，然后传地址给master。
5. master通知reduce任务的workers这些位置，然后用RPC读取map的workers这些数据（直接从map传到reduce workers？）。当reduceWorker读完所有中间数据后，按照键排序，以便将所有相同键的项组合。（排序是必须的，如果中间数据量大到内存不够了，用外部排序）
6. reduceWorker遍历已排序的中间数据，按照键值和对应的一组中间值传递给Reduce函数。reduce函数的结果加到此分区的输出文件中
7. 所有map和reduce任务完成后，master唤醒user程序。这时user程序中的MapReduce来return结果，一般所有分区输出文件同时输出传递给另一个MapReduce函数，不需要合并为一个文件。

### 数据结构：

每个map和reduce任务，保留存储状态（空闲，进行，完成）和工作计算机的标识（非空闲）

master是map任务中间文件位置的接受管道，master存储R个中间文件的位置和大小，map任务完成后，master接受信息更新，并且递增地传给正在进行的reduceWorkers

### 容错性：

#### workers failure

master定期ping每个worker如果有worker没有响应 或者 工作线程故障 。就标记为失败，任务重置为初始空闲状态，可以放在其他works上调度。

1. 对于map任务故障，直接全部重新执行，因为输出存储在故障计算机上，无法访问。

2. 对于reduce任务，只要接下来换机器执行，因为输出存储在全局文件系统中。

A执行map任务的时候坏了转给了B，那么对应执行reduce的workers需要重新执行，从B读取数据

对于故障 需要有很大的弹性。

#### master failure

写入master数据结构的检查点，master终止之后，可以从上一个检查点状态重新启动新的副本。

中止计算，客户端可以检查这种情况，重试

#### 存在故障的语言（没看懂）

map和reduce函数确定后，分布式产生的输出与无错误的顺序执行产生的输出相同。

主要是通过map和reduce输出的原子提交实现这一属性。reduce任务输出文件生成一个，map任务生成R个。map任务完成后，workers向master发送一条消息，包含R个文件名。reduce任务完成后，自动文件重命名为最终输出文件（如果多个worker执行相同的reduce任务，则相同的最终文件多个重命名调用）。

如果map和reduce运算是确定的，等同顺序执行，比较容易逻辑推理；但是如果存在非确定的运算，在存在非确定性运算符的情况下，特定归约任务R1的输出等同于由非确定性程序的顺序执行产生的R1的输出。然而，不同归约任务R2的输出可以对应于由非确定性程序的不同顺序执行产生的R2的输出。

举例子：M的一次输出可能被R1读取并正在执行，M的另一次输出可能被R2读取了

#### 本地化

如果失败将在输入数据的副本附近调度map任务，大多数是本地读取的，不会消耗带宽

#### 任务粒度

任务数量M和R远大于计算机数量（当然有个上限），调整动态负载平衡。 R受到用户限制，一般是机器数量的小倍数

主机必须做出O(M+R)调度决策并在内存中保持O(M∗R)状态？？？

#### 任务备份

问题：有些机器运行落后，最后几个任务要一直等着。

解决方案： 当一个MapReduce系统快完成的时候，主线程调度正在进行的任务进行备份执行，只要主执行或者备份执行有完成的，那这个任务就被标记完成。这样加了少量计算负担，能提升时间较多

### 优化拓展

#### 分区函数

用户指定M和R，使用分区函数对数据进行分区。要求平衡的分区，有时候希望相同的输出文件结束。。。

#### 序列化

保证给定分区中，中间KV按照K升序处理，方便输出文件被查找。

#### 组合函数

中间键有时候有大量重复计算，原本都需要单独发给reduce任务计算，然后合并。

现在可以在map任务的计算器上指定一个合并函数，执行合并后写入一个中间文件，发给reduce任务。这样可以提高操作速度。

这种部分合并的combiner函数和reduce函数代码一致（除了输出的位置）。

#### 输入输出类型

支持不同格式读取输入数据，读取器不一定需要提供从文件读取的数据

#### 跳过不良记录

用户代码的错误导致崩溃，修复错误 忽略(跳过)错误  检测记录确定性崩溃，发送信号给主程序，下次重新执行就能跳过记号

本地调试

状态信息：状态页面查看所有任务的进度

计数器工具：对各种事件记录，在任务函数中调用incre，定期传给master以更新

## GO

方法是带着接受者参数的函数，接受者需要是包内定义的类型。带指针参数的函数必须接受一个指针，但是以指针为接收者的方法被调用时，接收者既能为值也能为指针。反过来也是对的，接受值参数的函数必须接受值，但是接收者既能为值也能为指针。

chan ：默认情况下，发送和接收操作在另一端准备好之前都会阻塞。这使得 Go 程可以在没有显式的锁或竞态变量的情况下进行同步

go 函数名(参数列表)

### 接口

接口：定义为接口后，把其他类型的实例赋值给他，然后可以对应的不同方法，来实现多态

接口是一种抽象的类型，有点像泛型

```go
// 接口不关你是什么类型，只要你实现了这个方法
// 定义一个类型，一个抽象的类型，只要实现了say()就可以称为sayer类型
type sayer interface(){
    say(参数) 返回值列表
    say2()
    say3()
}
// 自定义类型，实现了say方法
type cat struct{}
func (c cat) say(){
    fmt.Println("喵喵喵")
}
type dog struct{}
func (d dog) say(){
    fmt.Println("汪汪汪")
}
type person struct{}
func (p person) say(){
    fmt.Println("嗷嗷嗷")
}
// 接口类型的入口，interface作为入参类型
func da(arg sayer){
    arg.say()
}
// 主函数
func main(){
    c1 := cat{}
    da(c1)
    d1 := dog{}
    da(d1)
    p1 := person{}
    da(p1)
}
```



## lab1

```shell
go run mrsequential.go wc.so pg*.tx // 之后的两项是传给mrsequential的命令行参数，分别是一个动态库和所有电子书

文件wc.go以及mrapps目录下的其它几个文件，都定义了名为map, reduce的函数，这两个函数在mrsequential.go中加载并调用。给mrsequential绑定不同的*.so文件，也就会加载不同的map, reduce函数。如此实现某种程度上的动态绑定。

我们的代码主要写在src/mr目录下的几个文件，这几个文件由src/main目录下两个文件mrcoordinator.go, mrworker.go调用。这两个文件的作用是启动进程、加载map, reduce动态库，并进入定义在src/mr目录下的主流程
```

### 1. RPC和chan实现数据交换

#### PRC

实现worker与cooridinator之间**rpc交互** call函数通过端口，调用rpcname（Coordinator.函数名），远程调用（Coordinator.函数名）这个函数，然后通过内存地址读取取回结果

目的 ： RPC 约等于PC

客户端有一个函数，服务器端有一个函数的实现 ； RPC保证远程调用，传递参数，然后返回结果

过程： 在 RPC 中，客户端和服务器端都有一个 stub。Stub 是指一份代理，它代表了本地对象与远程对象之间的一个代理，实现了客户端和服务器端的交互。当客户端调用 stub 上的方法时，stub 内部会将请求（包括函数的类型、参数等）封装成一条消息并发送给服务器端，（中间数据有序列化和反序列化）服务器端接收到请求并将执行结果返回给 stub，然后 stub 再将执行结果返回给客户端（中间数据有序列化和反序列化）。这个过程中，stub 将客户端和服务器端的交互细节封装了起来，让客户端和服务器端像本地调用一样进行 RPC 调用，从而降低了复杂度和使用难度，提高了开发效率。

在客户端和服务器端的 stub 中，通常包含了一些调用远程方法所需的信息，如方法名、方法参数等，这些信息可以通过特定的协议进行传输。在传输层面，RPC 中比较常见的协议有 TCP、HTTP 和 HTTPS 等。在使用这些协议进行远程方法调用时，stub 还可以负责序列化和反序列化请求和响应消息，这些消息通常使用类似 JSON、XML 或二进制编码形式进行表示。

```go
 // 如果不需要参数和返回值，最好设置为空的结构体，不建议设置nil
    var reply int // 初始值为0
    err := client.Call("Arith.Ping", &struct{}{}, &reply)
    if err != nil {
        panic(err)
    }
```



##### 用法例子

下面是一个使用 Go 实现 RPC 调用的例子：

**server.go**

```go
package main

import (
    "errors"
    "net"
    "net/rpc"
    "time"
)

type Args struct {
    A, B int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
    time.Sleep(1 * time.Second) // 模拟耗时操作
    *reply = args.A * args.B
    return nil
}

func main() {
    arith := new(Arith)
    rpc.Register(arith)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", ":5555")
    if err != nil {
        panic(err)
    }
    defer l.Close()
    for {
        conn, err := l.Accept()
        if err != nil {
            continue
        }
        go rpc.ServeConn(conn)
    }
}
```

这里定义了一个 `Arith` 类型，并为它实现了一个 `Multiply` 方法。`Multiply` 方法接收 `Args` 类型的参数，将其相乘后保存在 `reply` 指针指向的整型变量中。

程序启动后，会监听在 5555 端口，并等待客户端的 RPC 调用请求。

**client.go**

```go
package main

import (
    "fmt"
    "net/rpc"
    "time"
)

type Args struct {
    A, B int
}

func main() {
    client, err := rpc.DialHTTP("tcp", "localhost:5555")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    args := &Args{7, 8}
    var reply int
    start := time.Now()
    err = client.Call("Arith.Multiply", args, &reply) // call stub生成消息 远程调用 

    if err != nil {
        panic(err)
    }
    fmt.Printf("Arith: %d*%d=%d,time: %v\n", args.A, args.B, reply, time.Since(start))
}
```

这里定义了一个Args结构体类型, 使用rpc.DialHTTP来与RPC服务端建立连接, 然后向服务端发送一个 Multiply 方法的 RPC 调用，入参是 args，返回值是 reply。最后，将服务端返回的结果打印出来。

当程序运行时，客户端会向服务端发送一个 args.A=7,args.B=8 的请求，服务端在收到请求后会执行 Multiply 方法，将返回值 56 填充到 reply 变量中并返回给客户端。最后客户端打印出这个结果。

当服务端执行 Multiply 方法时，还用了一条 `time.Sleep(1 * time.Second)` 语句来模拟一个耗时操作。这样，我们就可以看到整个 RPC 调用过程的耗时了。



#### chan

channel 是golang特有的类型化消息的队列，可以通过它们发送类型化的数据在协程之间通信，可以避开所有内存共享导致的坑；通道的通信方式保证了同步性。数据通过通道：同一时间只有一个协程可以访问数据：所以不会出现数据竞争，设计如此。数据的归属（可以读写数据的能力）被传递。

```go
var ch chan Type // 定义类型为 Type 的 chan
ch <- data // 发送数据 data 到 chan ch
data := <- ch // 从 chan ch 中接收数据，并将其赋值给变量 data   接收操作会一直阻塞，直到有数据可以被接收。如果不希望接收操作被阻塞，可以使用带缓冲的 chan
var ch = make(chan Type, capacity) // 带有缓冲区的 chan 容量为 capacity  capacity 表示 chan 的缓冲区大小，可以在创建 chan 的时候指定。带缓冲的 chan 可以在没有读取者时发送数据，也可以在没有写入者时接收数据，但是如果缓冲区已满或为空时，发送和接收操作仍然会被阻塞

```

此外，chan 还可以使用 select 语句实现多路复用，同时处理多个 chan 的读写操作，类似于 Linux 中的 select() 系统调用。在 select 语句中，可以通过 case 语句处理 chan 的读写操作，default 语句用于处理没有可读和可写的 chan 的情况。当有多个 case 同时满足条件时，程序会随机选择一个 case 进行处理。例如：

```go
select {
  case data := <- ch1:
    // 处理从 ch1 中接收到的数据
  case ch2 <- data:
    // 处理将数据 data 发送到 ch2 中
  default:
    // 没有任何 chan 可读也没有可写的数据
}
```



##### 用法例子

下面是一个简单的 Go 程序，使用 chan 实现了两个协程之间的通信。在这个程序中，有一个 jobs chan 用于发送工作任务，一个 results chan 用于接收工作结果。有3个协程(worker)同时从 jobs chan 中接收任务，完成任务后将结果发送到 results chan 中。在 main 函数中，我们先向 jobs chan 发送了5个任务，然后通过接收 results chan 中的结果来等待工作的完成。这样，就实现了两个协程之间简单的通信和同步。

```go
package main

import (
    "fmt"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
    // <-chan是receive-only channel而 chan<-是 send-only channel
    // 在 Go 中，如果将一个 chan 赋值给另一个 chan 变量时，就会复制一个新的 chan,而且读写的属性一致
    // 在 Go 中，一个 chan 可以同时被多个协程读或写，但会有竞争写入或者竞争读取的问题，需要通过锁或者其他同步原语来解决。为了保证并发安全，建议在使用 chan 时，尽量将读写 chan 的操作限制在一个协程中，每个 chan 只由一个协程读或写。
    for j := range jobs {
        fmt.Printf("worker %d started job %d\n", id, j)
        time.Sleep(time.Second)
        fmt.Printf("worker %d finished job %d\n", id, j)
        results <- j * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // 启动三个协程进行工作
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // 发送 5 个任务到 jobs chan 中
    for i := 1; i <= 5; i++ {
        jobs <- i // 主协程写一个，协程池就会拿一个，chan并发安全 保证只有一个协程拿到运行，是线程安全的
    }
    close(jobs)

    // 接收工作结果
    for r := 1; r <= 5; r++ {
        <-results
    }
}
```

1. 我的做法是涉及的worker和cooridinator交互都用RPC交互，（其实部分用chan也可以吧）；
2. task相关信息队列的用chan就可以，用来分发task，天然安全，不会race

### 2. map和reduce处理

map 处理，用ihash处理下key分成Nreduce份用json编码后写出到"mr-x-y"文件。注意mr论文这步是有排序的，因为真正生产活动数据量是非常巨大的，map端提前排序好后，reduce的排序压力会减小很多。这里排不排序无所谓

map 处理结果 返回kv，存储在中间文件中，中间文件命名 mr-X-Y X是map任务号，y是reduce任务号；之后reduce再都读取中间文件；

因为真正分布式worker都不在一个机器上，涉及**网络传输**，所以用**json编码解码走个过场**

#### 设计结果：

map阶段应该将中间键划分为**nReduce**的存储桶，这里只用了一倍reduce数量也就是**%NumReduce**；每个映射器都应创建NumReduce个中间文件，存在main目录下以供Reduce任务使用。

Worker实现应该将第X个Reduce任务的输出放入文件MR-OUT-X中。Done()方法，在MapReduceJob完全完成时返回TRUE；此时，mrcoherator.go将退出。当作业完全完成时，工作进程应该退出。

#### Json编码

纯字符串形式的数据，轻量级的数据交换格式，可以跨平台、跨语言使用。

[JSON（JavaScript Object Notation）是一种轻量级的数据交换格式，它是基于 JavaScript 的一个子集，采用完全独立于编程语言的格式来表示数据，可以跨语言、跨平台使用。简洁清晰的层次结构使得 JSON 逐渐替代了 XML，成为了最理想的数据交换格式，广泛应用于 Web 开发领域](http://c.biancheng.net/json/what-is-json.html)[1](http://c.biancheng.net/json/what-is-json.html)[2](http://www.json.org/json-zh.html)。

JOSN编码和解码

```go
				enc := json.NewEncoder(tmp_file) // 创建JSON编辑器，用于写入tmp_file
				err = enc.Encode(bucket[i])     // 将buckt[i]的值编码为JSON格式，写入文件
```

```go
				for {
					var kv []KeyValue
					if err := dec.Decode(&kv); err != nil { //解码
						break
					}
					intermediate = append(intermediate, kv...) // ...表示把切片打散成元素后加入
				}
```

### 3 加锁

由于我用了channel所以在任务分配队列实现了天然并发安全，但是在别的地方还是遇到了问题，比如Done函数通过mrcoordinator主线程去时不时读取coordinator的状态来判断是否结束死循环。还有在一个worker调coordinator拉取数据的时候，另一个worker调coordinator的checkJobDone()函数进行检查。因此在响应可能发生冲突的地方加锁

### 4. 等待

1. GetTask调用失败，原因未处理work等待；
   worker有时候需要等待,比如当map任务都分发出去了,有的worker完成后又来申请任务，此时还有map未完成,reduce不能开始，这个worker需要等待下

解决：每次worker工作后都time.Sleep(time.Second)

### 5.crash test

1.  crash test失败，原因未处理可能损坏的worker
   mrapps/crash.go 随机干掉map reduce,看crash.go的代码是有一定几率让worker直接退出或者长时间延迟,可以用来测试恢复功能。这个逻辑是整合在map reduce函数里面的，注意worker被干掉时候任务已经拿到手了

解决：添加TimeTick函数，放到Done()函数中，检查map task 和 reduce task的时间，如果超过10秒，重新加载该Task

```go
				c.MapTask <- Task{FileName: c.files[i], MapId: i} //TODO 这里硬加Task没有去掉原来的Task 可能会导致跑很多次
/* task生命周期分析： 
		maptask在MakeCoordinator创建，定时间； 在做完在TaskFin设置true ； 在TimeTick检查是否超时，如果超时 重新添加这个任务
		这里其实判断超时没有区分是创建时间超时还是做任务超时，所以在做任务的时候设置时间
```

要记住go 的range第一个参数是index 

 strconv.Itoa 转化格式 字符串

os.Open 默认以只读方式打开，返回一个error； os.OpenFile可以设置打开文件的模式

## GFS

### 高容错的实现难点： 

1. 高性能的要求 -> 跨机器对数据分片
2. 大量计算机   ->  会有很多崩溃
3. 容错解决奔溃情况 -> 副本 
4. 副本 -> 需要实施更新，保持副本一致
5. 一致性高 -> 性能开销

### 一致性实现的难点 ：

1. 并发性 ，保证线程安全，可以通过之后的分布式锁等机制解决
2. 故障和失败案例 ：为了容错会用复制，但是不成熟的复制会导致读者在不做修改的情况下读到两次不同数据。

### GFS简介

 GFS旨在保持高性能，且有复制、容错机制，但很难保持一致性，现在谷歌已经不用了； 在论文中可以看到mapper从GFS系统(上千个磁盘)能够以超过10000MB/s的速度读取数据。论文发表时，当时单个磁盘的读取速度大概是30MB/s，一般在几十MB/s左右。

GFS的几个主要特征：

- Big：large data set，巨大的数据集
- Fast：automatic sharding，自动分片到多个磁盘
- Gloal：all apps see same files，所有应用程序从GFS读取数据时看到相同的文件（一致性）
- Fault tolerance：automic，尽可能自动地采取一些容错恢复操作

### GFS的Master的工作

- 维护文件名到块句柄数组的映射(file name => chunk handles)  **filename -> chunk handles -> chunk server \  version**

  这些信息大多数存放在内存中，所以Master可以快速响应客户端Client

- 维护每个块句柄(chunk handle)的版本(version)

- 维护块存储服务器列表(list of chunk servers)

  - 主服务器(primary)
    - Master还需维护每一个主服务器(primary)的**租赁时间(lease time)**
  - 次要服务器(secondaries)

  典型配置即将chunk存储到3台服务器上

- log+check point：通过日志和检查点机制维护文件系统。所有变更操作会先在log中记录，后续才响应Client。这样即使Master崩溃/故障，重启时也能通过log恢复状态。master会定期创建自己状态的检查点，落到持久性存储上，重启/恢复状态时只需重放log中最后一个check point检查点之后的所有操作，所以恢复也很快。

 这里需要思考的是，哪些数据需要放到稳定的存储中(比如磁盘)？

- 比如file name => chunk hanles的映射，平时已经在内存中存储了，还有必要存在稳定的存储中吗？

  需要，否则崩溃后恢复时，内存数据丢失，master无法索引某个具体的文件，相当于丢失了文件。

- chunk handle 到 存放chunk的服务器列表，这一层映射关系，master需要稳定存储吗？

  不需要，master重启时会要求其他存储chunk数据的服务器说明自己维护的chunk handles数据。这里master只需要内存中维护即可。同样的，主服务器(primary)、次要服务器(secondaries)、主服务器(primary)的租赁时间(lease time)也都只需要在内存中即可。

- chunk handle的version版本号信息呢，master需要稳定存储吗？

  需要。否则master崩溃重启时，master无法区分哪些chunk server存储的chunk最新的。比如可能有服务器存储的chunk version是14，由于网络问题，该服务器还没有拿到最新version 15的数据，master必须能够区分哪些server有最新version的chunk。

### GFS数据读取流程

![image-20230411161308484](D:\MyTxt\typoraPhoto\image-20230411161308484.png)

GFS通过Master管理文件系统的元数据等信息，其他Client只能往GFS写入或读取数据。

应用并发的通过GFS Client读取数据时，单个读取的大致流程如下：

1. Client向Master发起读数据请求
2. Master查询需要读取的数据对应的目录等信息，汇总文件块访问句柄、这些文件块所在的服务器节点信息给Client（大文件通常被拆分成多个块Chunk存放到不同服务器上，单个Chunk很大， 这里是64MB）
3. Client得知需要读取的Chunk的信息后，直接和拥有这些Chunk的服务器网络通信传输Chunks

### GFS文件读取

1. Client向Master发请求，要求读取X文件的Y偏移量的数据
2. Master回复Client，X文件Y偏移量相关的块句柄、块服务器列表、版本号(chunk handle, list of chunk servers, version)
3. Client 缓存cache块服务器列表(list of chunk servers)  
   ？ 只有一个Master，所以需要尽量减少Client和Server之间的通信次数，缓冲减少交互
4. Client从最近的服务器请求chunk数据(reads from closest servers)
   ？ 因为这样在宛如拓扑结构的网络中可以最大限度地减少网络流量(mininize network traffic)，提高整体系统的吞吐量
5. 被Client访问的chunk server检查version，version正确则返回数据
   ？ 为了尽量避免客户端读到过时数据的情况。

### GFS文件写入

 主要是关注文件写入的append操作，在mapreduce 中，reduce处理后计算结果需要append到file中

<img src="D:\MyTxt\typoraPhoto\image-20230411165609018.png" alt="image-20230411165609018" style="zoom: 67%;" />

1. Client向Master发出请求，查询应该往哪里写入filename对应的文件。

2. Master查询filename到chunk handle映射关系的表，找到需要修改的chunk handle后，再查询chunk handle到**chunk server数组映射关系的表**，以list of chunk servers(主副本primary、secondaries、version信息)**也就是得到所有该文件的chunkserver并且附加上版本号**作为Client请求的响应结果

   接下去有两种情况，已有primary和没有primary(假设这是系统刚启动后不久，还没有primary)

   - 有primary，（**primary：主文件副本，也就是正常情况下对外读写的文件；secondaries： 其他副本**）

     继续后续流程

   - 无primary

     - master在chunk servers中选出一个作为primary，其余的chunk server作为secondaries

       (暂时不考虑选出的细节和步骤)

       - master会增加version（每次有**新的primary**时，都需要考虑时进入了一个new epoch，所以需要**维护新的version**），然后向primary和secondaries发送新的version，并且会发给primary有效期限的租约lease。这里primary和secondaries需要将version存储到磁盘，否则重启后内存数据丢失，无法让master信服自己拥有最新version的数据(同理Master也是将version存储在磁盘中)。

3. Client发送数据到想写入的chunk servers(primary和secondaries)，有趣的是，**这里Client只需访问最近的secondary，而这个被访问的secondary会将数据也转发到列表中的下一个chunk server**，**此时数据还不会真正被chunk severs存储**。（即上图中间黑色粗箭头，secondary收到数据后，马上将数据推送到其他本次需要写的chunk server）
   ？ 这么做提高了Client的吞吐量，避免Client本身需要消耗大量网络接口资源往primary和多个secondaries都发送数据。

4. 数据传递完毕后，Client向primary发送一个message，表明本次为append操作。

   primary此时需要做几件事：

   1. primary此时会检查version，如果version不匹配，那么Client的操作会被拒绝
   2. primary检查**lease是否还有效**，如果自己的lease无效了，则不再接受任何mutation operations（因为租约无效时，外部可能已经存在一个新的primary了）
   3. 如果version、lease都有效，那么primary会选择一个offset用于写入
   4. primary将前面接收到的数据写入稳定存储中

5. primary发送消息到secondaries，表示需要将之前接收的数据写入指定的offset（更新其他副本）

6. secondaries写入数据到primary指定的offset中，并回应primary已完成数据写入

7. primary回应Client，你想append追加的数据已完成写入

### GFS一致性

1. appen一次之后，read返回什么结果

   例如： M（maseter），P（primary），S（Secondary）

   某时刻Mping不到P了，那咋办？ 

   假设选出新P，那可能旧P还在和CLient交互，两个P同时存在叫做脑裂，会出现严重问题，比如数据写入顺序混乱等问题，严重的脑裂问题可能会导致系统最后出现两个Master。

   所以依照M知道的旧P的lease，在旧Please期限结束之前，不会选出新P，也就是M和P无法通信，但是P还可能和其他Client通信

2. 强一致性的保证
   保证 所有S都成功写入或者都不写入

GFS是为了运行mapreduce而设计的

## VMWare-FT

### The Design of a Practical System for Fault-Tolerant Virtual Machines

(容错虚拟机)[[https://blog.csdn.net/weixin_40910614/article/details/117995014]]

只针对一个backup 单核 一个storage

### 备份实现方法

实现分布式容错常用方法是 **主/备份方法**；一致性的保证需要保持同步，两种方法： 

1. 状态转移，主服务器持续发送搜索状态变化给备份服务器，带宽需求大
2. 备份状态机，备份状态机视为确定状态机（DSM）。主/备份服务器相同的初始状态启动，然后相同的顺序接受并执行相同的输入请求，得到相同的输出，这样保持一致性，方法复杂，数据传输量小很多。
3. 状态转移是记录checkpoint点，同步完后再返回给client；复制状态机是操作记录同步后，再响应client

采用备份状态机会有一个问题，即对于真实物理机来讲，有些行为是不确定的（例如，中断，产生的随机数等），所以很难将行为看作是一个确定性（deterministic）的行为。
VMWare的解决方案是将所有操作虚拟化，并在VM和物理机之间加一个Hypervisor（虚拟机管理程序），然后就可以通过hypervisor让VM都能模拟成一个DSM运行，那么primary的所有操作都确定了，再通过Logging channel将这些操作传递到backup，然后backup进行replay，从而实现两者的高度一致。
具体是怎么将不确定的行为进行确定化操作的，它主要是进行了一个写日志的操作来保证的

### 复制状态机具体复制的是什么操作

应用程序级别： 比如GFS的文件append或write，需要关注复制状态机如何处理的，一般需要修改应用程序本身

机器层面操作： 状态state是寄存器的状态、内存的状态，操作operation则是传统的计算机指令；这种情况复制状态机感知的是底层机器指令，用硬件很昂贵，这里用的是虚拟机实现的

<img src="D:\MyTxt\typoraPhoto\image-20230413151512850.png" alt="image-20230413151512850" style="zoom: 50%;" />

### 非确定信息导致差异

非确定性的指令，网络包接受/时钟中断，多核（不讨论）

VM-FT对非确定性指令会先拦截，然后模拟执行，并且同步

### 失败场景处理

primary操作完成后，和client断了，这时候backup接手，那这次操作就失效了。

实际上不会发生这个错误，primary响应client之前，会通过logging channel发送消息给backup，backup相同操作后发ack，primary收到ack后，才响应client

raft也是一样

### 基本架构

1. 避免脑裂，对外只有一个primary输入输出，通过logging channel传递给Backup达到同步
2. 两个虚拟机部署，一个磁盘，一个网络
3. 确定性存放：也就是做一个确定状态机，具体是**Hypervisor**操作虚拟化后，把所有primary的确定、不确定（比如随机生成一个数）事件都记录log entry流，传给backup，保证一致性
4. FT Protocol：每次primary把output**写到日志给backup后才发送**，目的是保证backup随时能接手且**一致**
   保证**故障转移**（当primary失败时，我们希望切换到backup提供服务）
5. FT Logging Buffers and Channel： primary和backup两端的日志缓存差距保持在一个范围，用到了生产者消费者模型
6. 检测 故障响应：primary和backup两端的UDP心跳检测、logging channel检测
7. 虚拟机恢复FT VMotion：对VM进行克隆

### 与GFS比较

1. VM-FT备份的是**计算**，GFS备份的是存储。
2. VM-FT提供了相对严谨的一致性，并且对客户端和服务器都是透明的。可以用它为任何已有的网络服务器提供容错。例如，可以将 FT 应用于一个已有的邮件服务器并为其提供容错性。
3. GFS 只针对一种简单的服务提供容错性，它的备份策略会比 FT 更为高效。例如，GFS 不需要让所有的操作都应用在所有的 Replica 上。
4. VM-FT**不适合进行高吞吐量的服务**，性能下降很多，毕竟它要对整个系统内的内存、磁盘、中断等都进行 replay，这是一个巨大的负担

## Raft

### 复制状态机

复制状态机（Replicated State Machine，简称RSM）是一种分布式系统的设计模式。在该模式下，一个服务或应用程序的状态机被复制到多个节点上并进行并行处理，以提高可用性和性能。

具体地说，采用复制状态机的系统中，对于某个特定的客户端请求，它将被发送到所有复制的状态机或者其中的一组。然后，每个状态机独立地执行相同的操作序列，并生成相同的结果。最终，生成的结果将会被汇总并返回给客户端。

通过这种方式，复制状态机可以将单点故障风险降至最低并提高系统的可靠性。此外，由于并行执行相同的操作序列，该模式还能够提供更好的性能和可伸缩性。

复制状态机是一种常见的分布式系统设计模式，在诸如Google、Facebook和Amazon等互联网巨头公司的分布式系统中得到广泛应用。

### 单点故障

前面介绍过的复制系统，都存在单点故障问题(single point of failure)。

- mapreduce中的cordinator
- GFS的master
- VM-FT的test-and-set存储服务器storage

 而上述的方案中，采用单机管理而不是采用多实例/多机器的原因，是为了避免**脑裂(split-brain)**问题。

 不过大多数情况下，单点故障是可以接受的，因为单机故障率显著比多机出现一台故障的概率低，并且重启单机以恢复工作的成本也相对较低，只需要容忍一小段时间的重启恢复工作。

**为什么单机管理能避免脑裂问题** ？

比如有两个strorage，要选出primary，那可能有网络分区的原因，storage两个分区产生两个primary，对外界的client来说有两种不同的

### 大多数原则 majority rule

如何解决脑裂： 如果投票得大于一半，多数的那个成为leader，也就是多数的分区会继续运行，如果没有多数 系统不能运行

### 用Raft构造复制状态机RSM

这里raft就像一个library应用包。假设我们通过raft协议构造了一个由3台机器组成的K/V存储系统。

系统正常工作时，大致流程如下：

- Client向3台机器中作为leader的机器发查询请求
- leader机器将接收到的请求记录到底层raft的顺序log中
- 当前leader的raft将顺序log中尾部新增的log记录通过网络同步到其他2台机器
- 其他两台K/V机器的raft成功追加log记录到自己的顺序log中后，回应leader一个ACK，
  （如果有一台网络问题比较慢，那也达到了majority）
- leader的raft得知其他机器成功将log存储到各自的storage后，将log操作反映给自己的K/V应用
  （网慢的那台这时候才发ACK，）
- K/V应用实际进行K/V查询，并且将结果响应给Client

系统出现异常时，发生如下事件：

- Client向leader请求
- leader向其他2台机器同步log并且获得ACK
- leader准备响应时突然宕机，无法响应Client
- 其他2台机器重新选举出其中1台作为新的leader
- Client请求超时或失败，重新发起请求，**系统内部failover故障转移**，所以这次Client请求到的是新leader
- 新leader同样记录log并且同步log到另一台机器获取到ACK
- 新leader响应Client

### Raft的log

用途：持久化、顺序化 操作数据 ，方便重传，方面查看同步操作进行的情况

格式：

有很多的log entry（入口），比如log index 、 leader term这些唯一标识；每个log entry 有 command 和 leader term信息；

保证最后机器的所有log的一致性

### Raft的选举

followers收不到leaders的心跳，election time超时，开始重新选举，新的leader term

这时候就算原leader接进来，也会发现是新的leader term， 不会唱脑裂

#### 防止选举死循环

如果两个followers的election time 几乎同时到齐，都成为candidate，那会一直竞争leader死循环;

一般采用election time为随机值，防止同时发起选举

#### 选举超时时间

略大于心跳时间，太短会频繁选举，而且选举过程中是对外宕机，太长检测不到，本身也是宕机

加入些随机数，防止分裂选举死循环

 Raft论文进行了大量实验，以得到250ms～300ms这个在它们系统中的合理值作为eleciton timeout。

## FT-Raft-2

### 选举规则

处理别节点发来的RequestVote RPC时，需要检查限制，满足以下之一才能赞同票：

1. 候选人最后一条Log条目的任期号**大于**本地最后一条Log条目的任期号；或者
2. 候选人最后一条Log条目的任期号**等于**本地最后一条Log条目的任期号，且候选人的Log记录长度**大于等于**本地Log记录的长度

成为leader的限制：

1. 大多数原则
2. 当选的机器一定是具有最新的term的机器

### 日志覆写同步（未优化版本）

所有raft节点维护两个Index

* nextIndex数组：乐观的变量，所有raft节点都维护`nextIndex[followerId]`用于记录leader认为followerId的下一个需要填充log的index。
  更新时间：每个leader当选之后都会乐观的认为所有的follower的nextIndex是自己的log最后的下一个，而在实际appendEntries或者installSnapshot的时候如果发现日志同步没有那么乐观就会根据情况减小next，把之前没有同步的log先同步上。

* matchIndex数组：悲观变量，，所有raft节点都维护`matchIndex[followerId]`用于记录leader认为followerId的已经确认的log最后一个的index，表示在此之前的log都是和当前的leader确认一致的。
  更新时间：每个leader当选之后都会悲观的认为自己已经确认过所有的follower的nextIndex是0，因为还没开始确认。

未优化版本的问题：Raft集群中有出现log落后很多的server，leader需要进行很多次请求才能将其log与自己对齐

### 日志擦除

新上任的leader发现其他follower有和自己不一样的log，就会把从那个log开始往后的logs都擦除，来保证leader的日志权威性一致性

### 日志快速覆写同步（优化）

落后较多的log（比如新机器接入、宕机恢复很久）要同步到现在leader一致的话，按照之前的逐步回退很慢，浪费网络资源

优化后的log catch up quickly过程：

|      | logIndex1 | logIndex2 | logIndex3 | logIndex4 | logIndex5 |
| ---- | --------- | --------- | --------- | --------- | --------- |
| S1   | term4     | 5         | 5         | 5         | 5         |
| S2   | term4     | 6         | 6         | 6         | 6         |

1. S2的leader是term7当选，那nextIndex = 6，发送hearbeat随带log是(空, 6, 5)，意思是(当前nextIndex指向的term，nextIndex-1的term, nextIndex-1的值)
2. S1收到心跳，对比自己的logIndex为term5，与之前不同的是，除了no顺带回复自己的log信息(5,2), 意思是(请求中logIndex位置的值，当前值最早出现的logIndex位置)
3. S2收到回应后，把nextIndex改为2，下次附带([6,6,6,6],4,1), 意思是nextIndex之后的数据是[6,6,6,6]
4. S1收到后，检查logIndex1是term4对齐了，更新一致性

广播的时候，面对大量followers， 用每个go程单独处理followers

### 持久化

就是把一些全局的变量（比如currentTerm）写到（持久化存储）磁盘里，当然操作回复之前写入，类似于同步过程

持久化原因： 如果重启之后不知道任期号，很难确保任期只有一个leader

一个Raft节点崩溃重启（和新加入节点一样）后，必须重新加入

除了重新加入重新执行本地的log，更偏向于快速重启，上次持久化快照位置开始，这就要考虑持久化一些状态量：

- vote for：投票情况，因为需要保证每轮term每个server只能投票一次
- log：崩溃前的log记录，**因为我们需要保证(promise)已发生的(commit)不会被回退**。否则崩溃重启后，可能发生一些奇怪的事情，比如client先前的请求又重新生效一次，导致某个K/V被覆盖成旧值之类的。
- current term：崩溃前的当前term值。因为选举(election)需要用到，用于投票和拉票流程，并且**需要保证单调递增**(monotonic increasing)
- lastIncludedIndex: 奔溃前的快照保存的最后一个log的index，到这个log位置都是保存在快照中的，也就是状态机的持久化数据，如果崩溃后重启就不用交这个log之前的log了，因为上层可以直接读取整个快照。
- lasIncludedIndex: 奔溃前的快照保存的最后一个log的term

### 服务恢复

类似的，服务重启恢复时有两种策略：

1. 日志重放(replay log)：理论上将log中的记录全部重放一遍，能得到和之前一致的工作状态。这一般来说是很昂贵的策略，特别是工作数年的服务，从头开始执行一遍log，耗时难以估量。所以一般人们不会考虑策略1。
2. **周期性快照(periodic snapshots)**：假设在i的位置创建了快照，那么可以裁剪log，只保留i往后的log。此时重启后可以通过snapshot快照先快速恢复到某个时刻的状态，然后后续可以再通过log catch up或其他手段，将log同步到最新状态。（一般来说周期性的快照不会落后最新版本太多，所以恢复工作要少得多）

 这里可以扩展考虑一些场景，比如Raft集群中加入新的follower时，可以让leader将自己的snapshot传递给follower，帮助follower快速同步到近期的状态，尽管可能还是有些落后最新版本，但是根据后续log catch up等机制可以帮助follower随后快速跟进到最新版本log。

 使用快照时，需要注意几点：

- 需要拒绝旧版本的快照：有可能收到的snapshot比当前服务状态还老
- 需要保持快照后的log数据：在加载快照时，如果有新log产生，需要保证加载快照后这些新产生的log能够能到保留

### 使用Raft

 重新回顾一下服务使用Raft的大致流程

1. 应用程序中集成Raft相关的library包
2. 应用程序接收Client请求
3. 应用程序调用Raft的start函数/方法
4. 下层Raft进行log同步等流程
5. Raft通过apply channel向上层应用反应执行完成
6. 应用程序响应Client

- 并且前面提过，可能作为leader的Raft所在服务器宕机，所以Client必须维护server列表来切换请求的目标server为新的leader服务器。
- 同时，有时候请求会失败，或者Raft底层失败，导致重复请求，而我们需要有手段辨别重复的请求。通常可以在get、put请求上加上请求id或其他标识来区分每个请求。一般维护这些请求id的服务，被称为clerk。提供服务的应用程序通过clerk维护每个请求对应的id，以及一些集群信息。

## Lab2

<img src="D:\MyTxt\typoraPhoto\image-20230508170955517.png" alt="image-20230508170955517" style="zoom:33%;" />

![image-20230508171059069](D:\MyTxt\typoraPhoto\image-20230508171059069.png)

Raft 复制状态机协议

实现FT ，难点是故障导致副本一致性受损，难点是调试

把client请求组织到log序列汇总，确保所有副本服务器看到相同的Log，来保证一致性；如果服务器故障之后回复，Raft会将日志更新；所以只要大多数服务器活着且能通信，Raft就继续，不然就宕机直到大多数机器能通信才开始从中断位置通信

Raft作为关联方法的Go对象类型，作为模块。一组RAFT实例通过RPC相互通信以维护复制的日志。您的RAFT界面将支持不确定的编号命令序列，也称为日志条目。这些条目用索引号进行编号。具有给定索引的日志条目最终将被提交。此时，您的RAFT应该将日志条目发送到更大的服务以供其执行。

### 时间

三个时间的比较： broadcastTime 100« electionTimeout 200-400« 平均故障时间

两个RPC ：  AppendEntriesRPCs是 leader进行日志复制和心跳时使用的 << RequestVoteRPCs是候选在这选举过程中使用的

两个时间驱动用单独的goroutine驱动 

简单的实现方式是在raft中维护一个变量，包含leader的最后一次消息的时间，并用goroutine用sleep()驱动定期检查是否超时

定时的设置想了很久，把elctiontime作为下一个选举的时间点，定期检查是否超过这个时间点，如果超过了就要选举，heartbeattime设置为时间长度，用sleep相应长度然后发送heartbeat	

先设置 100 和 150-300，一次心跳不到重选

### log

- `entry`：Raft 中，将每一个事件都称为一个 entry，每一个 entry 都有一个表明它在 log 中位置的 index（之所以从 1 开始是为了方便 `prevLogIndex` 从 0 开始）。只有 leader 可以创建 entry。entry 的内容为 `<term, index, cmd>`，其中 cmd 是可以应用到状态机的操作。在 raft 组大部分节点都接收这条 entry 后，entry 可以被称为是 committed 的。
- `logs`：由 entry 构成的数组，只有 leader 可以改变其他节点的 log。 entry 总是先被 leader 添加进本地的 log 数组中去，然后才发起共识请求，获得 quorum 同意后才会被 leader 提交给状态机。follower 只能从 leader 获取新日志和当前的 commitIndex，然后应用对应的 entry 到自己的状态机

### election

candida并发请求选票，大于一半的时候成为leader，保证只成为一次 用sync.Once类型

这是Go编程语言中的代码，`sync.Once`是Go标准库中的一种并发原语，用于确保某个函数只会被执行一次。

在该代码中，`becomeLeader`是一个变量，使用`sync.Once`作为其类型，在程序运行过程中，我们可以通过调用`becomeLeader.Do(func)`的方式来确保`func`只会被执行一次。每当我们需要执行`func`时，我们可以多次调用`becomeLeader.Do(func)`函数，但实际上，该函数只会执行一次。

### RPC的处理

每个RPC应该在自己的goroutine中发送并处理回复，因为：

1. 到不了的peer不会延迟收集选票的过程
2. electiontime和heartbeattime可以继续在任何时候计时

election是并行的给集群中的其它机器发送 RequestVote RPCs

选取投票的写法： 一个函数开始投票，分出很多go程每个单独对接，每个函数对接发RPC和后处理

### RequestVote

安全性的限制：Raft使用选举过程来保证一个候选者必须包含有所有已提交的日志才能胜出；请求中包含了 leader的日志信息，如
果投票者的日志比候选者的日志更新，那么它就拒绝投票 :

1. 候选人最后一条Log条目的任期号**大于**本地最后一条Log条目的任期号；或者
2. 候选人最后一条Log条目的任期号**等于**本地最后一条Log条目的任期号，且候选人的Log记录长度**大于等于**本地Log记录的长度

成为leader的限制：

1. 大多数原则
2. 当选的机器一定是具有最新的term的机器

### AppendEntries 

每次当 leader发送 AppendEntries RPCs请求的时候，请求中会包含当前nextIndex后面的日志记录 和 他直接前继的任期和索引，

1. 如果存在一条日志索引和 prevLogIndex相等，但是任期和 prevLogItem不相同的日志，需要删除这条日志及所有后继日志。

2. 如果 leader复制的日志本地没有，则直接追加存储。

以上两条需要分别进行，如果直接用leader发来的日志记录覆盖follower的日志，那会产生bug，因为这个RPC可能是过时的RPC，所以需要严格保存这两条的分别执行。

### applyCh

作为一个ApplyMsg的chan，日志提交之后 需要添加发送的条目

需要用单独的goroutine实现，因为可能阻塞，必须是一个不然难以确保按照日志顺序发送

go程用sync.Cond在不满足发送条件的时候等待, 需要提交的时候唤醒

### 日志同步

nextIndex是leader认为的下次发给其他follower新log的首位置，在当选的时候会自认为所有follower的nextIndex都和自己一样，是自己日志的记录的最后一条的+1位置；

commitIndex是每个server被提交日志后最新log的索引。

lastApplied是每个server提交状态机的最新log的索引。

**Raft 保证下列两个性质**：

- 如果在两个日志（节点）里，有两个 entry 拥有相同的 index 和 term，那么它们一定有相同的 cmd；
- 如果在两个日志（节点）里，有两个 entry 拥有相同的 index 和 term，那么它们前面的 entry 也一定相同。

通过”仅有 leader 可以生成 entry”来确保第一个性质， 第二个性质则通过一致性检查（consistency check）来保证，该检查包含几个步骤：

leader 在通过 AppendEntriesRPC 和 follower 通讯时，会带上上一块 entry 的信息， 而 follower 在收到后会对比自己的日志，如果发现这个 entry 的信息（index、term）和自己日志内的不符合，则会拒绝该请求。一旦 leader 发现有 follower 拒绝了请求，则会与该 follower 再进行一轮一致性检查， 找到双方最大的共识点，然后用 leader 的 entries 记录覆盖 follower 所有在最大共识点之后的数据。

寻找共识点时，leader 还是通过 AppendEntriesRPC 和 follower 进行一致性检查， 方法是发送再上一块的 entry， 如果 follower 依然拒绝，则 leader 再尝试发送更前面的一块，直到找到双方的共识点。 因为分歧发生的概率较低，而且一般很快能够得到纠正，所以这里的逐块确认一般不会造成性能问题。当然，在这里进行二分查找或者某些规则的查找可能也能够在理论上得到收益。

#### 边界条件考虑

初始存入 0 0 空作为第一个log，作为dummy节点。在添加快照功能之后需要把logs[0].index改成lastIncludedIndex。在这之前，保持dummy节点作为锚点不被改变

#### 快速同步

实现的主要方法是在reply中添加一个Xindex，每次RPC结束后都通知leader更新nextIndex[serverID]。

如果落后的较多，则返回Xindex为最后一个log的index + 1通知leader发送下次从这个Xindex开始日志串

如果index匹配了，但是term不一致，则按照论文给出的优化方法，找到这个term对应的第一个日志并放到Xindex中存起来，这样 leader
接受到响应后，就可以直接跳过所有冲突的日志（其中可能包含了一致的日志）。这样就可以减少寻找一致点的过程。

按照这个方案2B的catch up测试还是比较慢，用27秒，老师给出的是17秒

#### 一个RPC重试的BUG

原来的实现：

​	如果appendEntriesLeader发现follower的日志不同步，重新修改后递归appendEntriesLeader，也就是重发RPC直到同步完成。但是因为本身就是上锁的函数，递归调用自己，还要解锁再加锁，就算这样了递归调用不知道有错，错误不匹配后retry，但是第二次还出错不retry了。

修改后的实现：

​	如果同步失败，则修改nextIndex，等着下一次心跳或者快照来同步

### 测试结果说明

| 时间 | peer数量 | RPC数量 | PRC总字节数 | RPC中日志目录条数 |
| ---- | -------- | ------- | ----------- | ----------------- |
| 3.1  | 3        | 60      | 15190       | 0                 |



### 日志提交一致性

leader知道有log entry到达majority servers，他就会commit这个log。

有特殊情况是达到了majority servers但还没commit的leader crash，所以不通过判断majority servers来判断是不是该commit，因为就算达到了majority server之后的leader可能还会全部覆盖。Raft采用只有的currentTerm的leaders的logs被提交，因为选主过程中的**Log Matching Property**原则就已经保证了选出的leader在整个集群中日志的完整性。

#### applier日志

1. 如果上任leader在提交日志之前宕机，下一任 leader将尝试完成日志的复制。这时候，如果有rf.logs[%v].Term: %v != rf.currentTerm的情况，因为新的leader不能准确判断这个log是不是已经提交，就不能去提交这个log，也就是只提交当前term的日志；
2. 如果在上个term残留的日志后面有新的日志满足提交的条件，因为**Log Matching Property**已经保证了在此之前的日志都是一致性的，那就会把新日志之前的所有log都保证提交，顺便也提交了之前的term残留的log。

### 锁的使用

1. 考虑修改server的状态变量的时候一定要上锁，
2. 在可能需要wait的操作不要上锁：channel的读写，等待timer，sleep()，发送RPC ，重传RPC（已经删除）
3. 先考虑大粒度的锁，不要提前优化，不过大粒度的锁也要考虑死锁问题

Log Matching Property：

1. 如果两个日志的两条日志记录有相同的索引和任期，那么这两条日志记录中的命令必然也是相同的。
2. 如果两个日志的两条日志记录有相同的索引和任期，那么这两个日志中的前继日志记录也是相同的

### 2B-bug记录 

1. append RPC 追加日志成功后重发还是发了日志片段，可能是nextIndex还没改过来就发下一次RPC了：目前推测发的是心跳，也就是所谓的过时RPC
   * 已经解决，心跳发空的日志 
   * 4.26 改成心跳不能空日志， 把 start()的RPC发送取消，交给心跳
2. 旧leader拿到新的log后回来，新leader以为他的nextIndex比实际的小1，而且新leader没有收到新的log所以发的只是心跳。这时候心跳检查, 可以让旧的leaders下台成为follower，log的长度匹配一样，但是因为nextIndex比实际的小1,检查的是前一个，index和term是匹配的，之后准备要把和leader所有不一致的删除，缺失的补上，但是因为是心跳，发过来的entry是空的。
   * 目前解决方案，已经改为心跳和普通同步一致，也就是也包含日志并且把 start() 的提交去掉，等待心跳提交，来减少RPC的超发，因为心跳很密集，同步比较快
3. student guide ：如果选举超时，没有收到现任领导的AppendEntries RPC，也没有向候选人授予投票权：转换为候选人。和我的处理不一样，可能会选举比较慢

### persistence

1. leaderCommit() 的提交log需要检查和当前的term是否一致，如果一致则提交，如果不一致不管。
2. 当leader提交新的log，那么之前的log间接提交，因为 log Matching Property

难改的测试： Figure8C leader commit判断通过之后，还没有applier就断了或者applier了之后断了，这样一千次

1. 因为persist只保存voteFor 、 log 、 currentTerm，所以commitIndex和 lastAppliyIndex在重启的时候还是0，那之前提交过的log现在还需要重头开始重新提交一遍
2. 日志同步的第一次retry没有成功，没有进行下一次retry，很奇怪。应该是死锁，注意锁起来的函数递归调用可能会死锁
   改为 日志同步RPC不 retry，修改nextIndex交给下一次心跳同步

- vote for：投票情况，因为需要保证每轮term每个server只能投票一次
- log：崩溃前的log记录，**因为我们需要保证(promise)已发生的(commit)不会被回退**。否则崩溃重启后，可能发生一些奇怪的事情，比如client先前的请求又重新生效一次，导致某个K/V被覆盖成旧值之类的。（例如，leader的宕机重启，logs如果没有保存而丢失的话，重新请求一致性的log，相同的index可能会产生不同的logs，而旧的log如果已经apply了，那可能会破坏一致性）
- current term：崩溃前的当前term值。因为选举(election)需要用到，用于投票和拉票流程，并且**需要保证单调递增**(monotonic increasing)
- lastIncludedIndex: 奔溃前的快照保存的最后一个log的index，到这个log位置都是保存在快照中的，也就是状态机的持久化数据，如果崩溃后重启就不用交这个log之前的log了，因为上层可以直接读取整个快照。
- lasIncludedIndex: 奔溃前的快照保存的最后一个log的term

### 快照

用于日志压缩，快照之后，存储当前的状态，那么这个点之前的日志就可以删除。

![image-20230504201733607](D:\MyTxt\typoraPhoto\image-20230504201733607.png)

一般机器单独的进行快照，除非有一个很慢或者新加入的follower需要leader网络发送快照来使其快速追赶

InstallSnapshot RPC

leader发送快照RPC，followers来决定使用；

1. 如果快照包含新信息超过follower的logs, 那会完全选择快照覆盖和logs删减；
2. 如果快照比follower的logs短，那prefix覆盖，之后的保留

**快照和一致性的冲突的**：虽然违背了只有leader修改logs的强领导原则，但是快照的时候一致性已经达成了，所以没有决定是冲突的，数据流还是leader流向follower

快照的时机：确定日志总量达到一个阈值进行快照

### InstallSnapshot RPC

1. 快照RPC发送的时机：

   落后较多的follower在同步日志的catchup过程中，leader回退自己的Log，到起点无法回退，这很好leader发送快照给FollowerFollower，之后通过AppendEntries将后面的Log发给FollowerAppendEntries将后面的Log发给Follower

2. 之前的log没有applier怎么办，直接把快照包装成ApplyMsg，传到applyCh由上层raft处理
3. 日志发送的是到lastIncludedindex位置的前面的全部，没有按照论文所说的实现偏移量

### 压缩后修改index

因为压缩之后logs数组的下标就不是index了，需要大改。。。

1. lastLog的index可能变成了以一个dummy节点，index =0，需要改成nextIncludedIndex，因为在选举的时候需要判断

2. nextIndex 等等也需要改

3. 添加一个接口，然后一个个测，改掉所有的下标

   func (rf *Raft) GetRealLastLogIndex() int {

     return Max(rf.logs[len(rf.logs)-1].Index, rf.LastIncludedIndex)

   }

### 2d-快照bug

1. 先写所有server的快照功能，需要大改下标，主要是加了rf.GetRealLastLogIndex()函数

2. 快照之后，上一次append还没发出去的Entry会被部分覆盖，也就是说在go程去append 的时候args被其他线程修改了

这个bug改了好久，并发的问题，因为之前args.Entry的创建直接用数组切片初始化Entries:    rf.logs[nextindex-rf.LastIncludedIndex:]，可能编译器折叠args.Entry为切片的表达式，也就是并发过程中切片本身变化可能会导致args.Entry的变化？？

之后改成了make空间，然后用copy(args.Entries, rf.logs[nextindex-rf.LastIncludedIndex:])赋值，bug解决

3. crash后恢复的server需要把之前的log重新应用到状态机（这里重新应用logs应该采用提交快照应用），否则直接往后提交新追加的logs是无法提交的，因为lastApplied没有持久化是恢复的时候是0，用日志applier的方式会从log1开始交，但是lastIncludedindex之前的log已经被快照剪短了。

改成make的时候lastApplied和commitIndex都初始化为lastIncludedindex，但是这时候也可能会出现直接提交新的log，而没有实现重新应用之前的log到状态机。这里测试能通过，不清楚什么原因？？？？可能是之前的状态也持久化了？

真正的原因是在2D之后已经**持久化了快照**，那服务器重启的时候**上层状态机就可以直接读取硬盘中存储的快照**（之前的logs的总和操作），对于重启的server只需要继续添加之后的log就可以了。

### 优化

1. 因为持久化的操作在实际过程中是读写硬盘，比较花时间，实验原本给的persist.save()接口只能同时修改state和snapshot，之前在只修改state情况下采用的是保存一遍自己原本的snapshot。
   现在添加persist.saveStateonly()接口，去掉了没有新快照的时候重复读取自己的快照的操作

### 还存在的问题

1. 全部案例一般用时realt time460-470s, CPU时间 2.8s；加上-race 是八分4，CPU时间28s。
2. 没有竞争的全套实验2测试(2A+2B+2C+2D)所需的合理时间是6分钟的实时时间和1分钟的CPU时间。当使用-race运行时，它大约是10分钟的实时时间和2分钟的CPU时间
3. 主要是backup比知道是慢10s，2D不稳定

### git

文件夹有原来关联的仓库，git push不到自己的仓库。

```git
git status
//
git init
git remote add origin git@github.com:forthdifferential/6.824.git
git add .
git status

git add .
git commit -m "first commit"
git push
fatal: 当前分支 master 没有对应的上游分支。
为推送当前分支并建立与远程上游的跟踪，使用

    git push --set-upstream origin master

klchen@vmware:~/project/MIT6.824$ git push -u origin main
error: 源引用规格 main 没有匹配
error: 无法推送一些引用到 'git@github.com:forthdifferential/6.824.git'
klchen@vmware:~/project/MIT6.824$ git branch
* master
klchen@vmware:~/project/MIT6.824$ git push --set-upstream origin master

```

这之后上传的文件夹有个箭头，打不开，主要是因为链接了其他的仓库，也就是子文件夹中存在`.git`文件，导致冲突

解决方法： 

```git
1. 找到存在的.git文件然后删除，rm .git -r
2. git rm --cached [该文件夹名] git rm --cached 6.5840/
3. 重新执行git add .
4. 执行 git commit -m "rm old_git"
5. 执行git push origin [branch name]  git push
```

### 总结

1. raft集群一般是3 5 个，单数防止脑裂，一个服务器损坏的平均情况大概是几个月一次，所以3 5 个足够修复恢复了
2. 服务器遇到的问题会有网络分区联系不上、机器故障挂了等

有空看看Paxos，因为这是最经典的一致性算法

## Zookeeper

### 线性一致性

关于历史记录的定义，而不是关于系统的定义，也就是说这个系统对外界的请求表现出的是线性一致性。

计算机执行的序列，同时满足一下两个，就说明请求历史数据是线性的

1. 序列中的请求的顺序与实际时间匹配

2. 每个读请求看到的都是序列中前一个写请求写入的值

反之，如果序列的规则生成了一个带环的图，那么请求历史数据不是线性一致性的。

**重点是说**：1. 对于系统执行写请求，只能有一个顺序，所有客户端读到的数据的顺序，必须与系统执行写请求的顺序一致。

2. 对于读请求不允许返回旧的数据，只能返回最新的数据。或者说，对于读请求，线性一致系统只能返回最近一次完成的写请求写入的值。

这点在分布式系统中，不容易保证，因为副本很多，而在单个服务器的情况下天然保证。

* 一个符合线性一致性但是看起来回复不一致的情况：

客户端请求后一直没收到回复，重新发送请求，可能会受到上一次处理的结果，而不是最新的结果。
可能的原因是如果这是因为第一次系统执行完毕了并返回了，而响应故障或者回复报文网络丢包了，系统建立一个表存储为此客户端处理过的结果。

### Zookeeper介绍

Raft实际上就是一个库。你可以在一些更大的多副本系统中使用Raft库。但是Raft不是一个你可以直接交互的独立的服务，你必须要设计你自己的应用程序来与Raft库交互。

Zookeeper看成一个类似于Raft的多副本系统，运行在Zab（和Raft一样）之上。

为了提高系统的性能，需要把除了leader外的服务器用起来，但是线性一致性的要求表示只能通过leader交互，因为不能保证follower的日志是up-to-date。Zookeeper 的做法是不保证线性一致性，把读操作分摊给所有的peer处理，这说明请求一些落后的副本的时候可能会读取到一些旧的数据。

### Zookeeper的一致性保证

1. 保证写请求是线性一致性的，也就是执行是按照一致性的顺序；但是读一致性不考虑，不需要经过leader，可能会返回旧数据
2. 任何一个客户端的请求都会按照顺序执行（通常是看系统不一定会顺序执行的），FIFO，可能的实现是为所有请求打上编号，然后leader节点遵从这个顺序。

FIFO的实现注意：

1. 如果客户端在一个S读取，这个S挂了，切换到下一个S的时候，那从下一个S读取的数据必须要在上一次读取数据的log的后面。 也就是说第二个读请求至少要看到第一个读请求的状态！

​	  如果在这个位置的log之前，那新的S会阻塞客户端的响应或者告诉他找其他的S试试。

2. 对于单个的客户端，因为FIFO的要求，所以读写的线性一致性实际上能保证。比如一个写一个读，写给leader 17，读给了follower读，要阻塞到写完17才能读到17 ，不然客户端收到的回复很奇怪。

### Zookeeper的同步操作 （sync）

弥补非严格线性一致的方法是设计了一个操作类型是sync，效果相当于写请求。

如果我需要读最新的数据，我需要发送一个sync请求，之后再发送读请求。这个读请求可以保证看到sync对应的状态，所以可以合理的认为是最新的。但是同时也要认识到，这是一个代价很高的操作，因为我们现在将一个廉价的读操作转换成了一个耗费Leader时间的sync操作

### Zookeeper应用场景

如果你有一个大的数据中心，并且在数据中心内运行各种东西，比如说Web服务器，存储系统，MapReduce等等。你或许会想要再运行一个包含了5个或者7个副本的Zookeeper集群，因为它可以用在很多场景下。之后，你可以部署各种各样的服务，并且在设计中，让这些服务存储一些关键的状态到你的全局的Zookeeper集群中。

### Zookeeper中watch通知

大体上讲 ZooKeeper 实现的方式是通过客服端和服务端分别创建有观察者的信息列表。客户端调用 getData、exist 等接口时，首先将对应的 Watch 事件放到本地的 ZKWatchManager 中进行管理。服务端在接收到客户端的请求后根据请求类型判断是否含有 Watch 事件，并将对应事件放到 WatchManager 中进行管理。

在事件触发的时候服务端通过节点的路径信息查询相应的 Watch 事件通知给客户端，客户端在接收到通知后，首先查询本地的 ZKWatchManager 获得对应的 Watch 信息处理回调操作。这种设计不但实现了一个分布式环境下的观察者模式，而且通过将客户端和服务端各自处理 Watch 事件所需要的额外信息分别保存在两端，减少彼此通信的内容。大大提升了服务的处理性能。

需要注意的是客户端的 Watcher 机制是一次性的，触发后就会被删除。

### API

Znode类型

- regular：常规节点，它的容错复制了所有的东西
- ephemeral：临时节点，节点会自动消失。比如session消失或Znode有段时间没有传递heartbeat，则Zookeeper认为这个Znode到期，随后自动删除这个Znode节点
- sequential：顺序节点，它的名字和它的version有关，是在特定的znode下创建的，这些子节点在名字中带有序列号，且节点们按照序列号排序（序号递增）。

一些API用RPC调用暴露

### Ready file(znode)

Master节点在Zookeeper中维护了一个配置，这个配置对应了一些file（也就是znode）。通过这个配置，描述了有关分布式系统的一些信息，例如所有worker的IP地址，或者当前谁是Master。所以，现在Master在更新这个配置，同时，或许有大量的客户端需要读取相应的配置，并且需要发现配置的每一次变化 。

尽管配置被分割成了多个file，还需要保证有原子效果的更新：

如果Ready file存在，那么允许读这个配置。如果Ready file不存在，那么说明配置正在更新过程中，我们不应该读取配置。所以，如果Master要更新配置，那么第一件事情是删除Ready file。之后它会更新各个保存了配置的Zookeeper file（也就是znode），这里或许有很多的file。当所有组成配置的file都更新完成之后，Master会再次创建Ready file。

### Zookeeper实现计数器

处理并发的客户端请求，不能用单纯的读写数据库的方式，而且Zookeeper本身可能返回旧值，这也要考虑。

```
    WHILE TRUE:
    X, V = GETDATA("F")
    IF SETDATA("f", X + 1, V): // leader节点执行 ，要求读到最新的数据且版本号还没被修改
        BREAK
```

方式是min-transaction，mini版的事务的实现，把读-更改-写作为一个原子操作

## lab3

FT-kv存储，是复制状态机；

clien向k/v service发送三个RPC：Put Append Get，其中是Clerk 与 Client交互，Clerk通过RPC与servers交互。

Get/Put/Append 方法需要是线性的，对外表现是一致的，我理解就是说并发情况下 对外表现的线性一致性：

后一次请求必须看到前一次的执行后端状态；并发请求选择相同的执行顺序，避免不是最新状态回复客户端；在故障之后保留所有确认的客户端更新的方式恢复状态

### 3A-Key/value service without snapshots

每个kv service有一个raft peer，Clerks把Get/Put/Append  RPC发送给leader的kv service，进一步交给Raft，Raft日志保存这些操作，所有的kv service按顺序执行这些操作，应用到kv数据库，达到一致性

1. Clerk找leader所在kv service和重试RPC的过程；
2. 应用到状态机后leader通过响应RPC告知Clerk结果，如果操作失败（比如leader更换），报告错误，让他重试
3. kv service之间不能通信，只有raft peer之间RPC交互

### 构思：

主要过程： Clerk把操作包装Op发送 PRC给 所有server，如果server是leader就开始发给raft，raft同步完成后apply给server，server执行这个Op,返回给Clerk结果。

如果raft的applier超时，RPC回复超时，因为是旧leader，所以等到下一个term再发送RPC给新leader的server。

### 重复RPC检测

如果收不到RPC回复（no reply）,一种可能是server挂了，可能换一个重新请求；但是另一种是执行了但是 reply丢包了，这时候重新发的话会破坏线性一致性。 

解决方案是重复RPC检测：

clerk每次发RPC都发一个ID，一个Request一个ID，重发相同；在server中维护一个表记录ID对应的结果，提前检测是否处理过；Raft的日志中也要存这个ID，以便新的leader的表是正确的

如果之前的请求还没执行，那会重新start一个，然后等第一个执行完后表就有了，applCh得到第二个的时候看表再决定不执行了。也就是有两个一样的log 没关系的

### 请求表的设计

每个客户端一个条目，存着最后一次执行的RPC：

每个client同时只有一个未完成的RPC，每个client对PRC进行编号；也就是说当客户端发送第10的条目，那之前的都可以不要，因为之前的RPC都不会重发了；

执行完成后才更新条目，index是clientID，值是保留值和编号，

### client

每个client有一个clientID，是一个64位的随机值 

client发送PRC中有 clientID 和 rpcID ，重发的RPC序号相同



如果执行完被leaderID索引，内容仅包含序号和值

RPC处理程序首先检查表格，只有在序号>表格条目时才Start()s

每个日志条目必须包括客户端ID、序号

当操作出现在applyCH上时，更新client's table entry中的序号和值，唤醒正在等待的RPC处理程序(如果有)



### 3A-bug

1. 一个clerk就是一个client调用的，所以clientId应该在clerk结构里，之前我是 每个请求随机生成了一个clientId，这是不太合理的

2. 处理RPC超时不是 !ok表示的，应该要自己设置时间。本来使用的是条件变量等着表的更新，但是用select chanl可以用非阻塞处理处理不同的通道，而且可以直接传递变量，不需要重新查表了。 原代码

   ```cpp
   	if !isLeader {
   		reply.Success, reply.Err = false, "not leader"
   		return
   	}
   
   	for !kv.killed() {
   		if equset_entry, ok := kv.requsetTab[args.ClientID]; ok {
   			if args.RpcID <= equset_entry.rpcID {
   				// 已经执行RPC请求
   				reply.Success, reply.Value = true, equset_entry.value
   				return
   			}
   		} else {
   			DPrintf("SERVER {server %v}等待一致性通过,Optype: Get,ClientID: %v,RpcID: %v,Key %v", kv.me, args.ClientID, args.RpcID, args.Key)
   			kv.receiveVCond.Wait()
   		}
   	}
   ```

   定时器控制RPC超时 ，chanl传入处理完的数据

   疑问：定时器应该是一个函数一个还是放到server结构里，chanl也是一个函数一个，那开销好像比较大

   用 map[int]chan int处理 对应不同的index，提交时到对应chan中，index的唯一性。

   考虑添加chan和删除，delete(kv.chanMap, 1)   删除键为1的元素，自动销毁chan，考虑请求满足之后就会删除，同时存在chan也不多，先采用这种方案

2. 设计是定时器在server中，那server如果挂了，也是没有reply，不返回了，这种情况怎么解决？

这是服务器的问题，不是raft一致性没有达到，应该是RPC会返回false。

4. 看起来client认为的server编号和server自己编号不一致，把前者不打出log以免出错

   问题还有是重复检测发生在检查leader之前了，那可能不是leader但因为同步执行到了，也能返回，然后client更新了错误的leader。

   修改了原来的结构，改为先检查leader，然后重复检测，然后再start

5. **难改bug**TestSpeed3A 要求每个心跳周期至少处理三次client的请求，但是我的raft中，start收到新条目不会马上同步，而是等下一次心跳来同步。这个TestSpeed3A 是同一个客户端发送一千次append，RPC接受函数阻塞在那直到apply后回复，而我每次都是一个心跳处理一次，就会导致一个心跳周期值处理一个请求。

​		具体表现是：rpc一致性通过，执行完成，返回通道的时候超时了，然后一直重复超时

先改为start( )之后就发起心跳，试试

server处理操作的PRC超时时间设置有点不清楚，改到800ms这样大可以减小TestSpeed3A 不通过的概率，但是跑多了还是会有个错误。只可能原因是高并发多请求的时候apply处理不过来了

5.2 这中间还有一个bug，就是raft一致性达成了而且apply了太快了，通道还没make。那就不传这个通道，等超时返回，并在下一次RPC的重复检验的时候直接取到值

6. applCh拿到后发现表中已经有rpc，没有执行也没有返回，导致RPC得不到回复。

正确的做法是不执行，但是要返回值。查看rpc通道还存不存在，如果存在，说明上一次没有回复rpc（可能是上次chanl还没创建好）；如果不存在，说明这个条目是重复的条目，不需要返回RPC了

7. 好看点的写法是函数中间需要加锁的部分单独移出去做新函数，不然很丑而且容易忘记解锁
8. **严重bug** - leader的applyCh读取不到数据了，导致一直超时，一直操作不了数据库。 

原来是raft日志同步有问题，也就是在start后立即发起心跳引起的。

也就是修改后导致添加了额外的心跳，那在发送心跳的过程中snapshot了，会导致preLogIndex比follower的lastIncludeIndex还要小，这种情况我提前考虑，然后返回xindex为realLastlogIndex

9. ***改好久bug***leader 提交raft后没有达成了一致性但是没有apply上来，反倒是其余follower都apply上来了。

检查了一圈，问题应该是raft在同步完成后，leader检查可否更改commitIndex，如果通过更改commintIndex，那会用条件变量唤醒apply，但是实际上到这一步后没有成功唤醒apply。

原因是rf.applyVCond.Broadcast()唤醒不了在另一个go程上的applier，这种情况发送的不多，但是每次发送都是永远提交不上，原因也不知道。

我猜有一种可能是Broadcast()的时候，那个线程不在wait。所以把commitLeader()中唤醒的操作放到从之前的 每次增加commit唤醒 改到 函数最后再唤醒，这样减少applier压力。但是这样TestSpeed有时候过不了，还是改回原来，考虑其他优化

考虑没有进入wait状态就阻塞的原因，去找阻塞再哪一句，发现在写入chan的时候阻塞了。有可能是server一直在读导致的异常阻塞，把server改成select非阻塞读取。

另外chan写端阻塞可能是缓冲区的为0是非缓冲的（同步的），也就是只有读写双方都准备好了才可以进行，否则就一直阻塞。我把applyCh和map中的chan改为缓冲区是1，但是这只是缓兵之计，如果缓存超出，写端还是会阻塞

仔细想了下应该不是这个问题，可能是读端没读，还在跑之后的代码。

### 3B-Key/value service with snapshots

### 要求： 

测试程序将MaxraftState传递给您的StartKVServer()。

MaxraftState指示持久RAFT状态允许的最大大小(以字节为单位)(包括日志，但不包括快照)。**将MaxraftState与Persister.RaftStateSize()**进行比较。当您的键/值服务器检测到RAFT状态大小接近此阈值时，它应该通过调用RAFT的Snapshot来保存快照。

如果MaxraftState为-1，则不必创建快照。

MaxraftState应用于RAFT作为第一个参数传递给Persister.Save()的GOB编码的字节。

server重启的时候，从persister读取快照并且恢复

1. 快照需要存储的信息和快照的时机考虑清楚，kvserver的重复检验必须跨越检查点，所以任何状态必须包含在快照中。可能快照需要请求表因为一恢复就要启动重复检验的功能，应该是需要保存kv数据库，也就是状态机的状态
2. 快照中数据大写
3. raft可能被改错，记得常回看看

### 构思：

#### 快照时机：

考虑server和raft交互的时候，考虑server状态机和raft日志改变的时候。

在raft通过applyCh提交给server的时候，检查是否快照。快照的时候设置这个index之前，所以在应用状态机之前检查快照

#### 快照操作：

1. 检查stateSize，如果超出就做快照。
2. server做快照，需要保存数据库和请求表snapshot，保存index（因为之后还要传给让raft，且applyCh可能接受快照的，那时候需要比较快照的latest性）。然后把snapshot和index传给raft做快照，也就是保存raftstate
3. 恢复操作，server恢复状态机、请求表，raft恢复state
4. appCh可能会传条目，也可能传快照，分开处理。如果传的是快照，之前raftstate已经更新，判断快照合理性，应用快照。

### 3B-bug

1. raft提交一个snapshot并被server应用之后，又提交了一个相同index的Op，但内容是nil

原因的是server主动快照时在新applyCh传入条目的应用之前拍的，也就是传入快照函数的index是新条目的index-1而不是index 

2. 落后较多raft接受到RPC快照同步的时候，没有成功。仔细看原因是applCh写入snapshot的时候阻塞了。
2. raft下标越界，可能原因是用chan的时候解锁，锁被抢走了，修改了lastIncludeIndex,当chan写完获取锁的时候，已经改变了raft的状态。

对2 3bug的修改，主要是把安装快照的RPC中传给上层的chan放到最后处理，这时候PRC导致的状态更新已经完成了，chan放锁给其他的用来修改raft状态也没关系。

所以说chan不仅要记得解锁，还锁，而且要考虑在chan解锁期间发送的不确定性问题

## lab4

### 要求

在本实验中，您将构建一个键/值存储系统，该系统对一组副本组上的键进行“分片”或分区。分片是键/值对的子集；例如，所有以“a”开头的键可能是一个分片，所有以“b”开头的键都是另一个分片，等等。分片的原因是性能。每个副本组仅处理几个分片的放置和获取，并且这些组并行操作；因此，总系统吞吐量（每单位时间的输入和获取）与组数成比例增加。

您的分片键/值存储将有两个主要组件。首先，一组副本组。每个副本组负责分片的一个子集。副本由少数服务器组成，这些服务器使用 Raft 来复制组的分片。第二个组件是“分片控制器”。分片控制器决定哪个副本组应该为每个分片服务；此信息称为配置。配置随时间变化。

客户端咨询分片控制器以找到key的副本组，而副本组咨询控制器以找出要服务的分片。整个系统只有一个分片控制器，使用 Raft 作为容错服务实现。

分片存储系统必须能够在**副本组之间转移分片**。一个原因是一些组可能比其他组负载更多，因此需要移动分片来平衡负载。另一个原因是副本组可能会加入和离开系统：可能会添加新的副本组以增加容量，或者现有的副本组可能会因维修或退役而脱机。

本实验中的主要挑战将是**处理重新配置——将分片分配给组的变化**。在单个组内，所有组成员必须同意何时发生与客户端 Put/Append/Get 请求相关的重新配置。例如，Put 可能与重新配置同时到达，导致副本组不再对持有 Put 密钥的分片负责。组中的所有副本必须就 Put 是发生在重新配置之前还是之后达成一致。如果之前，Put 应该生效并且分片的新所有者将看到它的效果；如果之后，Put 将不会生效，客户端必须在新所有者处重新尝试。推荐的方法是让每个副本组使用 Raft 不仅记录 Puts、Appends 和 Gets 的顺序，还**记录重新配置的顺序**。您将需要确保在**任何时候最多有一个副本组为每个分片的请求提供服务**

重新配置还需要副本组之间的交互。例如，在配置 10 中，组 G1 可能负责分片 S1。在配置 11 中，组 G2 可能负责分片 S1。在从 10 到 11 的重新配置期间，G1 和 G2 必须使用 RPC 将分片 S1 的内容（键/值对）从 G1 移动到 G2。

> 注意：只有RPC 可以用于客户端和服务器之间的交互。例如，您的服务器的不同实例不允许共享 Go 变量或文件。

> 注意：本实验使用“配置”来指代分片到副本组的分配。这与 Raft 集群成员变化不同。您不必实施 Raft 集群成员更改。

本实验室的通用架构（一个配置服务和一组副本组）遵循与平面数据中心存储、BigTable、Spanner、FAWN、Apache HBase、Rosebud、Spinnaker 等相同的通用模式。不过，这些系统在许多细节上与本实验室不同，而且通常也更复杂、功能更强大。例如，实验室不会在每个 Raft 组中进化对等点集；它的数据和查询模型非常简单；分片的切换很慢，并且不允许并发客户端访问。

> 注意：您的 Lab 4 分片服务器、Lab 4 分片控制器和 Lab 3 kvraft 必须全部使用相同的 Raft 实现。

### 4A-The Shard controller

shardctrler 管理一系列编号的配置。每个配置都描述了一组副本组和分片到 replica groups的分配方案。每当此分配需要更改时，分片控制器都会使用新分配方案创建新配置。 Key/value clients and servers 在想要了解当前（或过去）配置时联系 shardctrler。

实现必须支持 shardctrler/common.go 中描述的 RPC 接口，它由 Join 、 Leave 、 Move 和 Query RPC 组成。这些 RPC 旨在允许管理员（和测试）控制 shardctrler：添加新的副本组、删除副本组以及在副本组之间移动分片。

1. 管理员使用 Join RPC 添加新的副本组。它的参数是map，从唯一的非零副本组标识符 (GID) 映射服务器名称列表数组。shardctrler 应该通过*创建*一个*包含新副本组的新配置*来做出反应。新配置应尽可能将分片**均匀**地分配到整组组中，并应移动尽可能少的分片以实现该目标。如果 GID 不是当前配置的一部分，shardctrler 应该允许重新使用它（即 GID 应该被允许加入，然后离开，然后再次加入）。

2. Leave RPC 的参数是以前加入的组的 GID数组。shardctrler 应该创建一个不包括这些组的新配置，并分配这些组的切片到其余组。新配置应该在组之间尽可能均匀地划分分片，并且应该移动尽可能少的分片以实现该目标

3. Move RPC 的参数是分片编号和 GID。 shardctrler 应该创建一个新的配置，其中指定分片被分配给指定组。 Move 的目的是让我们测试您的软件。 Move 之后的 Join 或 Leave 可能会undo Move，因为 Join 和 Leave 会重新平衡。

4. Query RPC 的参数是一个配置号。 shardctrler 回复具有该编号的配置。如果数字是 -1 或大于最大的已知配置数字，shardctrler 应该回复最新的配置。 Query(-1) 的结果应该反映 shardctrler 在收到 Query(-1) RPC 之前完成处理的每个 Join 、 Leave 或 Move RPC。

**第一个配置**应该编号为零。它不应包含任何组，并且所有分片都应分配给 GID 零（无效的 GID）。下一个配置（为响应 Join RPC 而创建）应编号为 1, &c。通常会有比组更多的分片（即，每个组将服务多个分片），以便可以以相当精细的粒度转移负载。

> 您的任务是在 shardctrler/ 目录下的 client.go 和 server.go 中实现上面指定的接口。您的 shardctrler 必须是容错的，使用来自实验 2/3 的 Raft 库。当您通过 shardctrler/ 中的所有测试时，您就完成了此任务。

提示：从您的 kvraft 服务器的精简副本开始。

提示：您应该为**分片控制器实现 RPC 的重复客户端请求检测**。 shardctrler 测试不会对此进行测试，但 shardkv 测试稍后会在不可靠的网络上使用您的 shardctrler；如果您的 shardctrler 没有过滤掉重复的 RPC，您可能无法通过 shardkv 测试。

提示：执行分片重新平衡的状态机中的代码需要是确定性的。在 Go 中，map迭代顺序是不确定的。(索引需要按照一定顺序遍历，不用range map)

提示：Go maps are references. If you assign one variable of type map to another, both variables refer to the same map. 因此，如果你想基于之前的 Config 创建一个新的 Config，你需要创建一个新的map对象（使用 make() ）并单独复制键和值。

提示：Go race detector (go test -race) 可以帮助你找到错误。

### shardctrler

shardctrler的client负责对其的Group和配置发送读或改的RPC请求。

1. 和lab3B差不多，就是raft实现的分布式server实现RPC访问，包括重复检验功能，RPC超时返回

初始分配，也就是confignum ==0 先考虑。

balance方法，输入一些副本组id（ []int），得到分配的情况（[NShards]int）

2. 新的map类型的Groups，先make，再按顺序存入原有gid对应的组，然后按顺序存入新的gid对用的组

得到新的分配方案的方法，需要原来方案的移动少，以及新方案的分配快

3. Leave RPC似乎只是测试用，所以先不做负载均衡重分配处理

### servers balance方案

1. 按照更新后的server组数计算平均数，以及余数，维护一个未分配的shards数组

2. 首先对原方案做调整，向新平均数靠近，回收多出来的shards存到未分配的数组中：

如果有余数，可以容许余数个Group的shard值是平均数+1，超出余数个后只能容许是平均数；

如果没有余数，只能容许所有的Group的shard值是平均值；

把以上两走情况结合，直接按照left>0判断，容许范围是average还是average+1

3. 最后对没有达到average的server增添，直到未维护shard分配完成。

因为之前以及保证了left个average+1的server，那么其他的server都是average个

#### 效果： 

1.除非必要（超出平均数或平均数+1），否则不移除原有的切片

2.对于不够的平均数的server，添加新的切片,达到average



#### bug

1. 之前的设计没有控制**left个数**的Group的shard值是平均数+1，而是直接放到了平均数+1，会导致以下分配缺陷：

对于不够的平均数的server，添加新的切片：

这时候可能会有个问题，就是在添加新切片的时候，未分配的shard的余量不够了，因为前面容许放到了平均数+1，可能出现对超出负载进行删减的过程中，没有删除任何平均数+1的server情况，那就没有未分配的shard给新的server。所以需要在删除的过程完成后，查看有没有剩下，如果没有剩下，就要删除到平均值，然后重新分配。但是这样修改之后会违背第一个性质，导致重分配浪费的情况。

修改为控制**left个数**的Group的shard值是平均数+1，代码设计时这对left变量修改即可

2. --- **FAIL**: TestMulti (0.63s)

     test_test.go:61: Shards wrong

   **FAIL**

   相同条件下配置需要时一模一样的，也就是gids数组是有一定排列顺序规则的。改为每次生成用于reBalance的gids后都sort以下

### 4B-Sharded Key/Value Server

#### 要求

每个 shardkv 服务器都作为副本组replica group的一部分运行。每个副本组为一些键空间分片提供 Get 、 Put 和 Append 操作。在 client.go 中使用 key2shard() 来查找密钥属于哪个分片。多个副本组合作为完整的分片集提供服务。 shardctrler 服务的**单个实例**将分片分配给副本组；当此分配发生变化时，副本组必须将分片传递给彼此，同时确保客户端不会看到不一致的响应。

您的存储系统必须为使用其客户端接口的应用程序提供**线性一致**的接口。也就是说，完成的应用程序调用 shardkv/client.go 中的 Clerk.Get() 、 Clerk.Put() 和 Clerk.Append() 方法必须显示为以**相同顺序影响所有副本**。 Clerk.Get() 应该看到由最近的 Put / Append 写入相同键的值，即使 Get和 Put与配置更改几乎同时到达。

只有当分片的 Raft 副本组中的**大多数服务器都处于活动状态**并且可以**相互通信**并且可以**与大多数 shardctrler 服务器通信**时，你的每个分片才需要取得进展。即使某些副本组中的少数服务器已死、暂时不可用或运行缓慢，您的实现也必须运行（服务请求并能够根据需要**重新配置**）。

一个 shardkv 服务器只是一个副本组的成员。给定某个副本组中的所有服务器永远不会改变（也就是一个raft集群为最小单位）。

我们为您提供 client.go 代码，该代码**将每个 RPC 发送到负责 RPC 密钥的副本组**。如果副本组表示它不负责这个key了，它会重试；在这种情况下，客户端代码向**分片控制器请求最新配置并重试**。作为**处理重复客户端 RPC** 支持的一部分，您必须修改 client.go，就像在 kvraft lab3中一样。

> 注意：您的服务器不应调用分片控制器的 Join() 处理程序。测试人员将在适当的时候调用 Join()。

> **Task**: 您的首要任务是通过第一个 shardkv 测试。在此测试中，只有一个分片分配，因此您的代码应该与 Lab 3 服务器的代码非常相似。最大的修改是让您的服务器检测**配置何时发生**并开始接受其key与它现在拥有的分片匹配的请求。

现在您的解决方案适用于静态分片情况，是时候解决配置更改问题了。您需要让您的服务器**监视配置更改**，并在检测到配置更改时启动分片迁移过程。如果一个副本组丢失了一个分片，它必须**立即停止向该分片中的key提供请求**，并**开始将该分片的数据迁移到接管所有权的副本组**。如果一个副本组获得一个分片，它需要**等待让前任发送全部的数据给它，然后才能接受该分片相关的请求**

> **Task:**在配置更改期间实施分片迁移。确保副本组中的所有服务器都在它们**执行的操作序列中的同一点执行迁移**，以便它们都接受或拒绝并发的客户端请求。// raft一致性通过再转移。
>
> 在进行后面的测试之前，您应该专注于通过第二个测试（“加入然后离开”）。当您通过所有测试（但不包括 TestDelete ）时，您就完成了此任务。

> 注意：您的服务器需要**定期轮询 shardctrler 以了解新配置**。测试预计您的代码大约每 100 毫秒轮询一次；更频繁是可以的，但太少可能会导致问题。

> 注意**：服务器需要相互发送 RPC，以便在配置更改期间传输分片**。 shardctrler 的 Config 结构包含服务器名称，但您需要 labrpc.ClientEnd 才能发送 RPC。您应该使用传递给 StartServer() 的 make_end() 函数将服务器名称转换为 ClientEnd 。 shardkv/client.go 包含执行此操作的代码。

提示：将代码添加到 server.go 以定期从 shardctrler 获取最新配置，并添加代码以在接收组不负责客户端密钥的分片时拒绝客户端请求。您仍然应该通过第一个测试。

提示：您的服务器应该向客户端 RPC 响应 ErrWrongGroup 错误,如果服务器不负责这个key（即其分片未分配给服务器组的密钥）。确保您的 Get 、 Put 和 Append 处理程序在并发**重新配置**时做出正确的决定。

提示：按顺序一次处理**一个**重新配置。

提示：如果测试失败，请检查 gob 错误（例如“gob：type not registered for interface ...”）。 Go 并不认为 gob 错误是致命的，尽管它们对实验室来说是致命的。

提示：您需要为跨分片移动的客户端请求提供至多一次语义（**重复检测**）。

提示：想想 shardkv 客户端和服务器应该如何处理 ErrWrongGroup 。如果客户端收到 ErrWrongGroup ，它是否应该更改序列号？如果服务器在执行 Get / Put 请求时返回 ErrWrongGroup，是否应该更新客户端状态？

提示：服务器迁移到新配置后，它可以**继续存储它不再拥有的分片**（尽管这在真实系统中会令人遗憾）。这可能有助于简化您的服务器实现。

提示：当组 G1 在配置更改期间需要来自 G2 的分片时，G2 在处理日志条目期间的什么时候将分片发送到 G1 是否重要？

提示：您可以在 RPC 请求或回复中**发送整个map**，这可能有助于简化分片传输的代码。

提示：如果您的某个 RPC 处理程序在其回复中包含作为服务器状态一部分的map（例如键/值map），您可能会因竞争而出现错误。 RPC 系统必须读取map才能将其发送给调用者，但它并没有持有覆盖map的锁。但是，您的服务器可能会在 RPC 系统读取它时继续修改同一个map。解决方案是让 **RPC 处理程序在回复中包含map的副本**。

提示：如果你在 Raft 日志条目中放置一个 a map or a slice ，你的键/值服务器随后在 applyCh 上看到条目并在你的键/值服务器的状态中保存对 a map or a slice 的引用，你可能有一个race。制作 map or slice 的副本，并将副本存储在键/值服务器的状态中。race是在你的键/值服务器修改 map/slice 和 Raft 读取它同时持久化它的日志之间进行的。

提示：在配置更改期间，两个组可能需要在它们之间**双向移动分片**。如果您看到死锁，这就是一个可能的来源。

### 构想

![image-20230529140159716](D:\MyTxt\typoraPhoto\image-20230529140159716.png)

#### 配置修改

* 什么时候修改配置

1. 100ms轮询sk配置，判断是否需要求改配置和数据迁移。
2. 配置每次只能+1的递增，且递增过程中产生数据迁移时不能修改配置。
3. 按照顺序一次处理一个重新配置，raft一致性通过后修改配置

* 配置修改互相之间怎么交换分片？

1. 建立一个新的RPC，处理groups之间的数据迁移
1. 采取pulling的方式，如果需要就向目标gid拉取，如果不再负责，则设置为outState为true给出

* 修改配置是否需要raft一致性通过？

1. 看Get请求，一开始就要判断这个分片是否处于当前的group，所以不是leader也需要知道是不可接受shard分片请求的情况，所以先设置raft一致通过后再修改配置，包括修改为不再负责

#### 配置修改的Op

* 还要考虑上述不是clientRPC操作（也就是处理配置shard）在raft实现后的**重复检查**吗？

  那就不用请求表检查，用这个操作有没有已经在执行来判断

* 用什么命名配置修改相关的Op的管道？

   还是raft的index 

* 配置修改的条目需要等待timeOut吗？

  timeOut是为了client请求的时候有个raft网络分区卡在那里不能取得一致性，也就是不再是真正的leader了，所以client需要换一个server来重新发送RPC;

  对于配置修改的Op，有可能这个raft也陷入网络分区，不再是leader就会导致一直卡着不能一致性后返回，所以还是得设置timeOut判断，timeOut之后不做为真正的leader，就不需要修改配置的操作了。等到下一次pollingConfig的goroutine判断为leader的server再重新请求配置

  timeOut是对于RPC设置的，RPC是单个goroutine。那我的修改配置也是开单独的goroutine。那么我就不需要像RPC那样一定要返回重试，改为定时重试，就不需要goroutine等着返回了

  所以完全不需要timeOut在内的返回通道

* 防止新配置应用之前添加更新的配置

决定配置修改后，也就是start新配置之后，就不能再pollingConfig；在配置修改apply，pulling新的shard完成，就重新开始可以

start一个原配置+1的新配置，还没改任何东西，这时候换leader，那新leader拉取的也是原配置+1的新配置；如果不换leader，检查configChecker的时候不会添加新的配置；如果原配置+1的新配置apply之后，需要立刻修改InState和OutState,然后再应用curconfig为新的config，如果配置没有变则可以立即开始下一次config检查，如果配置变了，还需要去pull shard；如果pull shard完成，InState改回false。

OutState不用于防止新配置应用之前添加更新的配置，而是给出shard的时候用

TOOD 之后可能OutState还在删除旧配置的时候，用于防止新配置应用之前添加更新的配置

#### 配置修改和同步新shard的顺序

同步新shard 肯定是在 配置修改同步完成并执行后来做的。 

同步完成修改InState为ture的shard就是需要向目标gid拉取的shard，检测到需要拉取在开始

* 是不是不能在更新配置完成后再去开goroutine拉取，否则如果更新配置完成后就宕机，那恢复之后也用于不回去拉取这个shard。

是，所以选择开一个新的groutine定期检查是否要拉取，这样重启后也能继续处理pullshard操作，而不会导致死锁。

这么做的依据是raft层的持久化保证了apply之后的条目是不会再修改或应用的（否则有可能一些旧值把新的修改了），也就是已经通过了，持久化后不需要再提交一次。导致不会进入pull shard的函数了

这里还要仔细想想raft持久化log和快照，恢复之后上层server具体是什么过程：

persist其实是raft层的定期保存，而snapshot是server层的定期保存。persist层更新更加频繁，为了保证一致性，而snapshot是为了server宕机重启后快速恢复保存的。

拿到num+1的配置，配置一致性通过，应用新配置，旧的等删，新的等拿到数据。

#### 客户端请求

1. 需要判断分片是不是在当前的gid中，如果不是先拒绝
2. 如果处在分片迁移过程中，迁移发出方需要立刻停止对该分片请求的控制（一迁移就修改为拒绝），接受方在迁移完成整个分片前停止对该分片的控制（全部迁移完再修改）

3. 重复检测的**客户端请求也需要按照分片处理**，因为数据迁移的时候，rpc的请求列表也是迁移的一部分，不然重复检测在迁移后起不到作用，相同的client的请求要重新开始记录。比如上一次是发给旧的group的请求，执行了但网络原因没有回复client，client重复请求的时候，当前向新的client发送请求，需要旧的server的请求表才能进行重复检测

请求表分片的效果是：分片的数据和它的请求表是一体的

#### Raft条目

1. 添加条目的之前检查分片管理，条目**一致性通过的时候再次检查分片管理**

如果一致性通过之后也需要判断分片管理情况，处理ErrWrongGroup的情况

2. group之间的采用pulling方式RPC拉取，但是单个group中在leader拉取之后还需要同步给所有raft实体，所以Op包含的数据应该要多点

包括**配置开始更改，开始删除shard，开始接受shard**都需要一致性通过。

问题：不能直接删除shard，因为可能需要的还没拿到shard，

#### Group之间的Pull shard

>  设置一个新的RPC互相交流，一直拉，直到拿到需要的shard并且一致化通过并应用

1. 按照inState为true的shard的，逐个shard拿

2. leader先拿到shard，然后raft同步shard到组内取
3. 一致性通过后，添加新的shard，修改inStats

* PRC接受函数方

1. 更新到互相交流的configNum一致时，才开始互相传递数据
1. 决定发送回给拿的group，也需要一致性通过，通过后如果是leader回复需要的shard，然后删除整个shard的数据，做垃圾清理。还要加上重复检测，如果RPC回复出现失败但已经执行的话，用重复检测。

被pullshard方的状态修改是需要一致性的，但是考虑了RPC返回时可能不成功，需要考虑重复请求，直到成功收到shard，应该有两种方案：

1. PullShard RPC接受函数进行重复检验，并且添加入raft层，等待一致性满足后，用chan返回leaderRPC返回。

2. PullShard RPC接受函数被拉取后，不修改，直到RPC请求方真正拿到了shard后，再增加一个RPC来告诉接收方，我已经拿到了shard，可以把那个shard的状态改过来，并且需要加入到raft层，等通过后修改状态，并删除整个shard的数据。这还要考虑另一个RPC的返回情况。

我选择1方案试试。方案1的重复检验比较麻烦，因为每次配置修改不仅拉取一次，按照分片拉取，而且增加了请求表，需要持久化和交换数据。

* 修改每次配置拉取按照group来拉，一次拉一个group，这样重复检测就是按照修改一次配置的一个请求顺序。

考虑一种情况，当被pull方的shard被拉去一次，raft一致通过后删除这个分片，但是因为网络原因RPC返回的时候断了，被shard已经被raft集群执行过了。对方没有收到需要的shard,下一次再来pullRPC的时候，重复检验，从请求表中读取这个shard的内容并且返回。如果这时候被pull方已经进入下一个配置，并且把这个shard删除，shard请求表也删除了,就不能实现重复的直接返回结果。这种情况，互相请求的请求表需要一直保存，纳入持久化

删除应该是要删除分片的数据，和分片的请求表，因为都发出去了。

> 看了很多人用方案2，也就是明确拿到了再删除shard对于的数据，这样删的更干净点。

1. server在pullshard完成后，shard改为wait状态，等待告知exower删除整个分片

2. 被pullshard后，shard状态先不改，保持out，因为还不能确定RPC发送方**真正拿到了shard并且一致化通过并应用**了,保持不能进入下一个config，等待删干净后进入下一个config check。

> 设置一个新的RPC，通知shard的ex-ower已经拿好并应用了，可以开始删除shard数据；
>
> 防止重启后不恢复RPC，开新线程来检测是否提醒ex-ower

delshards RPC告知ex-ower开始删除shards的数据，等一致化通过，返回chan告知已经删除了再返回RPC。如果RPC返回的时候断了，那因为已经执行了Del操作，看到数据删除了，直接返回成功就行。

* 这里和方案1的区别就是，执行了就直接返回OK，而方案1需要返回shard数据如果执行了，就只能一直卡着。

> PullShards PRC 和 DelShards RPC的重复检测
>
> 1. PullShards PRC：
>
>     args.NewConfigNum > kv.curConfig.NumRPC 接收方还没有更新到这个配置，让请求方再等等重新请求
>
>     args.NewConfigNum < kv.curConfig.Num 过时的RPC 舍弃这个RPC。因为被pull放只有 被通知可以删除整个shard后才会更新配置，所以被pull方配置更新则需求方已经拿到shard，说明只能是过时的RPC.
>
> 2. DelShards RPC：
>
>    因为DelShards RPC是类似于客户端请求，直接放到raft中，等apply后用chan回复。再次收到这个kv.curConfig更新不能保证接收方已经不在wait了。所以重复的RPC也需要返回OK（类似于客户端请求）。
>
>    考虑RPC失败的不确定性，就算回复是过时的RPC，也就是接收方已经删除了，发送方处理回复也需要添加一个Op到raft层。而是把Op的有效性判断，交给apply。
>
>    * 什么会出现args.ConfigNum > kv.curConfig.Num？
>
>      当G1宕机恢复的时候，正在走他曾经的人生，这时候curConfig比较小，其他在wait的server会发送最新的args.ConfigNum。这时候应该让其他server等一会，等G1归来之后，拉平ConfigNum，才能有效的接受RPC。
>
>    

#### 快照

快照测试过不了，重启之后，config变回了0，考虑是不是调整快照保存的变量

快照需要保存的变量：

还需要存curConfig、oldConfig否则恢复起来很慢

#### TODO

5.28-改配置，想好怎么改 //go kv.pollingConfig()开始设计





#### 重点关注的不确定

configchanging的取消位置

### bug

##### 5.30

1. 死锁，

   * 位置：leader提交新的config到raft层，一致化通过拿出后，应用新的config死锁。
   * 表现：leader修改配置的函数拿不到锁，但是follower能拿到锁
   * 解决：锁被configChecker()拿着，没有释放

2. RPC发给了未知的服务器，

   * 原因：第一次修改配置的时候，不需要pullRPC；好像此后num=2的时候还有这个问题，RPC发不出去
   * 解决：RPC名称写错了。。。

3. 案例跑不停

   配置修改后接受客户端的请求，还没有直接pull，是不是考虑用定时器，然后配置改完闹钟直接设置响起，直接拉取；

   否则来不及直接接受客户端的请求会来不及；先不改

4. 快照案例跑不停

   * 有的server拍摄快照之后，进度落后了，配置比别人慢就不能接受别人的pullshard请求。但是别人一直在发RPC请求pullshard，导致落后的server没空更新配置？？？
   * 分析：inStates和outStates需要都处理完全后才可以进行下一个配置，否则进入下一个配置后，其他的server来pullshard，因为配置不一致就不能发给他，而且配置不能倒退，永远不能让他pull走这段shard
   * 解决：添加outStates这部分的处理，在配置更改的时候改为需要被pull，控制不能修改config，被Pullshard拉走之后改为回来
   * 这里有个隐患：如果拉取了，但是RPC回复的时候网络断了，想再拉取就不好拉了，因为可能配置已经更新到下一个了。解决的方法：记录RPC，然后进行重复检验，如果被拉过了，那就把拉的部分重新给。（还没改，先等bug来）
   * 之后考虑，添加groups之间交流的RPC重复检验的功能

5. 诡异的bug，server一直在向自己pullshard，而且configNum落后很多，过去的自己一直在请求拉取现在的自己的部分shard，害怕。
   可能oldConfig的时候没有加锁，导致混乱了
   刚才再改这个诡异的bug被同学拍了一下，然后我直接吓死

6. pullshard的RPC设置为一直获取，直到拿到成功

   * 修改为返回Confignum不匹配就返回


#### 6.1

1. 修改groups之间传输分片为按照group传，可能一次穿多个分片，减少RPC发送；如果配置接收方配置还没跟上，先sleep100ms。
2. 出现有的group的curConfigNum = 0 的情况

考虑一种情况，如果shardkvService跑一段时间后，配置比较高，有一个kvGroup1死了，过了一会然后重新恢复。在G1宕机的时候，G2和G3可能需要从G1去pull shards，所以配置更新不了，也接受不了部分客户端请求。在G1恢复的时候，如果想要整个service继续运行，G1配置应该和之前的G2和G3保持一致，或者最多落后一个config。但是，如果没有快照，或者快照比较落后，恢复的配置很低，那S1想要catch up需要从G2和G3中pull shard，但是他们已经删除了。

如果G1恢复，没有快照，config从头开始。但是raft层持久化了，所有的logs都被保存，raft层重启的时候commitIndex和lastApplied重置为0，重新执行所有之前的logs。这些logs中包含了配置更新，添加新的shard，删除shard，nowait等。

* 恢复的效果：G1走过他的来时路，慢慢重新成为现在的他。
* 在这期间，三个独立goroutine的checker会往raft层加很多新的条目，再恢复期间加的这些条目都是无效的，因为index比较大，所以也不会对恢复过程产生不良影响。
* 直到正确返回客户端的请求后，就是他已经恢复完全了

#### 6.2

3. bug group宕机重启后，更新不了最新的配置，不能接受客户端请求
   * 分析： 101S2已经恢复 15726，猜测是正确处理完成测试了，考虑101S0。101的配置不是最新的，怎么没有更新新的配置？
   * 一般情况，配置更改指令加入raft层后，修改configchanging =1，当raft层执行完成后，应用这个条目修改配置，configchanging 再改回0。考虑一种情况，如果加入raft层后宕机了，重启的时候如果有快照恢复到快照的时刻，这时候configchanging 可能是1或者0。
   * 重启恢复的时候，可能会放进去新的config，修改configchanging为1，但是这个新的config不会被应用，也就是configchanging的值一直是0。
   * 问题一个是changing改为1后没有改过来0，    DPrintf("[Gid:%v]{S%v} applyConfigChange失败recvOp.Config.Num[%v] != kv.curConfig.Num[%v]+1", kv.gid, kv.me, recvOp.Config.Num, kv.curConfig.Num)
   
   [Gid:102]{S1} 重启的位置在10879，第一次卡住不让配置更新16156 
   
   configchanging这个参数没有一致化通过，宕机或者换leader的时候会突变。去掉这个控制，正常raft层加入新的配置，但是在apply 的时候控制配置的更新。
   
   configchanging删掉后bug解决。这个参数明显违背分布式原则。。。
   
4. 

[Gid:100]{S2}等着pull from[Gid:101]{S2}，但是目标service一直没有更新到这个config

[Gid:102]{S0}一直拉[Gid100]，但是拉不到。[Gid100]换leader了，

* NoWaitShards添加到raft层的操作，有一两条不是leader存的？

  可能失效的leader残留的发送RPC函数，增加判断leader操作以去掉

1. 还没恢复如初收到RPC，返回NoWaitShards ErrConfigNumNotMatch

5. 看起来有个group频繁的更换leader

   `[Gid:102]{S0} configChangeable Err,shardStates[9][In]`

​	`[Gid:102]{S2} configChangeable Err,shardStates[9][In]`

​	`[Gid:102]{S1} configChangeable Err,shardStates[9][In]`

可能是网络分区，都认为自己是leader，

恢复之后，有一个leader开始PullShards，向[Gid:100]{S1} 15358

[Gid:100]{S1}在拉取shardctrl配置的的时候newConfig.Num[0] <= curConfigNum[6]。为什么会请求一个0配置？

可能是shardctrl有问题传了个空的？

* 解决： 原来是shardctrl重复检测成功后直接返回了个空的config...............无语了

6. 发送Del RPC请求的那个server不再是leader，不能改回nowait

* 这时Del的PRC接收方等着删Out状态，发送方还是wait状态 
* [Gid:100]{S1}    [Gid:102]{S2}

#### 6.4

1. 修改pullShardsLeader等需要拉取配置的操作，为直接用kv的成员变量的，而不是Query，可以加速获取配置，而且试一下解决死锁。

* 除了定时的配置检查，需要向ShardkvCtrler拉取配置，其他情况都不向他拉取。 改完后终于稳定过了TestConcurrent2

2. bug-server保持在一些shardStates[OUT]状态，不能修改配置

   分析： [Gid:101]{S1} configChangeable Err,shardStates[4][Out]，configNum[4]

   5 [100 100 100 100 100 100 100 100 100 100]

   4 [100 100 100 100 100 102 102 102 102 102] 

   3 [100 100 100 100 101 102 102 102 101 101]

   2 [100 100 100 100 100 102 102 102 102 102]

   表现是：在configNum4的时候拍快照（在添加NoWaitShards，shards[[4]]之前），到configNum5后，宕机重新恢复，从configNum[4]开始恢复，读取lastIncludeIndex开始的日志。

[Gid:101]{S1} shard[4 8 9]改为StateOut，应用config4，拍快照1，[4 8 9]被PullShard完成，添加DelShards[8 9]到raft层，一致化通过，还没有正式apply这个DelShards操作，拍快照2，DelShards[8 9] apply了,DelShards[4]到raft层一致化通过，还没有正式apply这个DelShards操作，拍快照3,DelShards[4] apply了。Configchange到 5 到raft层一致化通过，快照4，Configchange到 5 apply了。

被kill，StartServer读取快照，但是好像没有恢复到快照4，而是配置5（不明原因）；

又一次重启：

设想：这次重启的时候，应该到快照4，可以看到raft层的日志只有一条，也就是，Configchange到 5 这一条

问题：但是这个第二次重启读取snapshot 的时候，好像回到了快照3的时候，并且没有日志重新apply上来，一直卡住



3. 看晕了，不想看了，之后测快照重启的冲突。和垃圾清理

#### 6.5

1. 重启恢复到快照的时候，之后继续同步很慢，可能是没有当前leader的term的日志，所以很多日志没有apply上来。

* leader 在当选时要先提交一条空日志，这样可以保证集群的可用性。这条空日志不能在raft层添加，
* 在server层开一个协程负责定时检测 raft 层的 leader 是否拥有当前 term 的日志，如果没有则提交一条空日志，这使得新 leader 的状态机能够迅速达到最新状态，从而避免多 raft 组间的活锁状态。
* 不能在raft层添加空日志，因为lab2可能过不了。raft层应该遵循严格的规则。

版本较低的 Group 在推进 Config 到这个版本之后已经正确处理过拉取数据或是清理数据的RPC也更新了 Shard 状态，但在重启后这最后处理的关键RPC对应的日志并没有重新被commit。原来，此时 leader 的 currentTerm 高于这个RPC对应的日志的 term，且这个时间节点客户端碰巧没有向该 Group 组执行读写请求，导致 leader 无法拥有当前任期的 term 的日志，无法将状态机更新到最新。



### 4Challenge

#### 状态垃圾收集

当副本组失去分片的所有权时，该副本组应从其数据库中删除丢失的键。保留它不再拥有且不再满足请求的价值是一种浪费。但是，这给迁移带来了一些问题。假设我们有两个组，G1 和 G2，并且有一个新的配置 C 将分片 S 从 G1 移动到 G2。如果 G1 在转换到 C 时从其数据库中删除 S 中的所有键，那么当 G2 尝试移动到 C 时，它如何获取 S 的数据？

使每个副本组保留旧分片的时间不要超过绝对必要的时间。即使副本组中的所有服务器（如上面的 G1）崩溃然后重新启动，您的解决方案也必须有效。如果您通过了 TestChallenge1Delete ，您就完成了这个挑战。

##### 实现

go kv.shardDelChecker()

定时检测已经拿到的shards，然后用DelShards RPC通知ex-ower可以删除整个shards相关的所有数据，然后把状态改为OK。

##### bug-添加删除shardRPC有些情况还是会超出

--- **FAIL**: TestChallenge1Delete (18.00s)

  test_test.go:803: snapshot + persisted Raft state are too big: 153019 > 117000

* 分析：这个测试的maxraftstate[1],也就是要求每次提交log后都做快照；比较Snapshot+raftState的大小；检查后发现raftState的大小有时候会出现9000+，考虑减少一些Op结构体中的数据，少向raft传点数据

* 2023/06/05 14:05:29 [gi0][n0]:raft[665],snap[15712]
  2023/06/05 14:05:29 [gi0][n1]:raft[602],snap[15712]
  2023/06/05 14:05:29 [gi0][n2]:raft[665],snap[15712]
  2023/06/05 14:05:29 [gi1][n0]:raft[597],snap[17734]
  2023/06/05 14:05:29 [gi1][n1]:raft[597],snap[17734]
  2023/06/05 14:05:29 [gi1][n2]:raft[597],snap[17734]
  2023/06/05 14:05:29 [gi2][n0]:raft[600],snap[15698]
  2023/06/05 14:05:29 [gi2][n1]:raft[662],snap[15698]
  2023/06/05 14:05:29 [gi2][n2]:raft[662],snap[15698]
  2023/06/05 14:05:29 结束检查大小total153079，expected117000

* 大小分布倒是挺均匀的，就是都超限了。可以看出主要是snap很大

* 1.把oldConfig不存，改为重启的时候查： 

  2023/06/05 14:18:14 [gi0][n0]:raft[602],snap[15598]
  2023/06/05 14:18:14 [gi0][n1]:raft[665],snap[15598]
  2023/06/05 14:18:14 [gi0][n2]:raft[665],snap[15598]
  2023/06/05 14:18:14 [gi1][n0]:raft[597],snap[17620]
  2023/06/05 14:18:14 [gi1][n1]:raft[597],snap[17620]
  2023/06/05 14:18:14 [gi1][n2]:raft[597],snap[17620]
  2023/06/05 14:18:14 [gi2][n0]:raft[600],snap[15584]
  2023/06/05 14:18:14 [gi2][n1]:raft[662],snap[15584]
  2023/06/05 14:18:14 [gi2][n2]:raft[600],snap[15584]

  稍微减了点，但没有用

* 2. 猜测可能是我是在apply应用log之前snapshot的，所以在最后一个log是delshard的时候我没有删除就snapshot，导致snapshot很大。预期可能是apply这个log之后再snapshot。

* 改到apply之后再snapshot检查后，解决这个最后的问题了！！结束了家人们

#### 配置更改期间的客户端请求

处理配置更改的最简单方法是在转换完成之前禁止所有客户端操作。虽然概念上很简单，但这种方法在生产级系统中并不可行；每当机器被带入或取出时，它会导致所有客户端长时间停顿。最好继续为不受正在进行的配置更改影响的分片提供服务。

修改您的解决方案，以便客户端对未受影响的分片中的键的操作在配置更改期间继续执行。当您通过 TestChallenge2Unaffected 时，您就完成了这个挑战。

虽然上面的优化很好，但我们仍然可以做得更好。假设某个副本组 G3 在转换到 C 时需要来自 G1 的分片 S1 和来自 G2 的分片 S2。我们真的希望 G3 在收到必要的状态后立即开始服务分片，即使它仍在等待其他分片。例如，如果 G1 关闭，一旦 G3 从 G2 接收到适当的数据，它仍应开始为 S2 提供服务请求，尽管到 C 的转换尚未完成。

修改您的解决方案，以便副本组在他们能够的时候开始服务分片，即使配置仍在进行中。当您通过 TestChallenge2Partial 时，您就完成了这个挑战。
